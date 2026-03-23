package httpserver

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/phenrril/tienda3d/internal/adapters/telegram"
	"github.com/phenrril/tienda3d/internal/domain"
)

var workshopClientSlugRe = regexp.MustCompile(`^[a-z0-9_]+$`)

func (s *Server) workshopAdmin(w http.ResponseWriter, r *http.Request) (*WorkshopAdmin, bool) {
	if !s.isAdminSession(r) {
		http.Redirect(w, r, "/admin/auth", http.StatusFound)
		return nil, false
	}
	if s.workshop == nil {
		http.Error(w, "módulo pedidos taller no disponible", http.StatusServiceUnavailable)
		return nil, false
	}
	return s.workshop, true
}

func parseWorkshopDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("fecha vacía")
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func parseMoneyAR(s string) (float64, error) {
	s = strings.TrimSpace(strings.ReplaceAll(s, ".", ""))
	s = strings.ReplaceAll(s, ",", ".")
	if s == "" {
		return 0, nil
	}
	return strconv.ParseFloat(s, 64)
}

func normalizeWorkshopClientSlug(raw string) (string, error) {
	s := strings.TrimSpace(strings.ToLower(raw))
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	if !workshopClientSlugRe.MatchString(s) {
		return "", fmt.Errorf("cliente debe ser snake_case (a-z, 0-9, _)")
	}
	return s, nil
}

func workshopStatuses() []domain.WorkshopOrderStatus {
	return []domain.WorkshopOrderStatus{
		domain.WorkshopPendiente,
		domain.WorkshopDisenado,
		domain.WorkshopEnImpresion,
		domain.WorkshopListoEntrega,
		domain.WorkshopEntregado,
	}
}

func (s *Server) renderAdminPedidosForm(w http.ResponseWriter, r *http.Request, edit bool, o *domain.WorkshopOrder, filaments []domain.WorkshopOrderFilament, errMsg string, _ int) {
	stock, _ := s.workshop.Filament.StockByColor(r.Context())
	data := map[string]any{
		"Edit":       edit,
		"Order":      o,
		"Filaments":  filaments,
		"Statuses":   workshopStatuses(),
		"Stock":      stock,
		"AdminToken": s.readAdminToken(r),
		"Error":      errMsg,
	}
	// Mantener status 200 para errores de formulario y renderizar la UI con popup
	// (evita que algunos clientes muestren HTML como texto plano).
	s.render(w, "admin_pedidos_form.html", data)
}

func parseFilamentsFromForm(r *http.Request) []domain.WorkshopOrderFilament {
	colors := r.Form["fil_color"]
	grams := r.Form["fil_grams"]
	var out []domain.WorkshopOrderFilament
	for i := range colors {
		c := strings.TrimSpace(strings.ToLower(colors[i]))
		if c == "" {
			continue
		}
		g := 0
		if i < len(grams) {
			g, _ = strconv.Atoi(strings.TrimSpace(grams[i]))
		}
		if g <= 0 {
			continue
		}
		out = append(out, domain.WorkshopOrderFilament{
			ID:        uuid.New(),
			ColorSlug: c,
			Grams:     g,
		})
	}
	return out
}

func sumFilamentsByColor(list []domain.WorkshopOrderFilament) map[string]int {
	out := make(map[string]int)
	for _, f := range list {
		color := strings.TrimSpace(strings.ToLower(f.ColorSlug))
		if color == "" || f.Grams <= 0 {
			continue
		}
		out[color] += f.Grams
	}
	return out
}

func (s *Server) validateStockForInPrint(ctx context.Context, oldFilaments, newFilaments []domain.WorkshopOrderFilament) (bool, string) {
	stock, err := s.workshop.Filament.StockByColor(ctx)
	if err != nil {
		return false, "No se pudo validar stock de filamento."
	}

	oldTotals := sumFilamentsByColor(oldFilaments)
	newTotals := sumFilamentsByColor(newFilaments)

	// Stock proyectado después de reemplazar filamentos del pedido.
	colors := make(map[string]struct{})
	for c := range oldTotals {
		colors[c] = struct{}{}
	}
	for c := range newTotals {
		colors[c] = struct{}{}
	}

	var deficits []string
	for color := range colors {
		projected := stock[color] + oldTotals[color] - newTotals[color]
		if projected < 0 {
			deficits = append(deficits, fmt.Sprintf("%s (%dg faltantes)", color, -projected))
		}
	}
	if len(deficits) == 0 {
		return true, ""
	}
	sort.Strings(deficits)
	return false, "No hay stock suficiente para pasar a en_impresion. Faltan: " + strings.Join(deficits, ", ")
}

func (s *Server) validateOrderCanEnterInPrint(ctx context.Context, o *domain.WorkshopOrder) (bool, string) {
	if o == nil {
		return false, "Pedido inválido."
	}
	required := sumFilamentsByColor(o.Filaments)
	if len(required) == 0 {
		return false, "Cargá al menos un filamento antes de pasar a en_impresion."
	}
	stock, err := s.workshop.Filament.StockByColor(ctx)
	if err != nil {
		return false, "No se pudo validar stock de filamento."
	}
	var deficits []string
	for color := range required {
		// Con el modelo actual, el consumo ya está asentado en ledger.
		// Para poder iniciar impresión, ese color no debe quedar en negativo.
		cur := stock[color]
		if cur < 0 {
			deficits = append(deficits, fmt.Sprintf("%s (%dg faltantes)", color, -cur))
			continue
		}
	}
	if len(deficits) == 0 {
		return true, ""
	}
	sort.Strings(deficits)
	return false, "No se puede pasar a en_impresion: " + strings.Join(deficits, ", ")
}

func workshopOrderFromForm(r *http.Request, existing *domain.WorkshopOrder) (*domain.WorkshopOrder, error) {
	slug, err := normalizeWorkshopClientSlug(r.FormValue("client_slug"))
	if err != nil {
		return nil, err
	}
	reqAt, err := parseWorkshopDate(r.FormValue("requested_at"))
	if err != nil {
		return nil, fmt.Errorf("fecha solicitud: %w", err)
	}
	delAt, err := parseWorkshopDate(r.FormValue("delivery_date"))
	if err != nil {
		return nil, fmt.Errorf("fecha entrega: %w", err)
	}
	detail := strings.TrimSpace(r.FormValue("detail"))
	isBarter := r.FormValue("is_barter") == "1" || r.FormValue("is_barter") == "on"
	var total *float64
	if !isBarter {
		t, err := parseMoneyAR(r.FormValue("total"))
		if err != nil {
			return nil, fmt.Errorf("total inválido")
		}
		if t <= 0 {
			return nil, fmt.Errorf("total debe ser mayor a 0 (o marcar CANJE)")
		}
		total = &t
	}
	st := domain.WorkshopOrderStatus(strings.TrimSpace(r.FormValue("status")))
	valid := false
	for _, v := range workshopStatuses() {
		if st == v {
			valid = true
			break
		}
	}
	if !valid {
		st = domain.WorkshopPendiente
	}
	var id uuid.UUID
	if existing != nil {
		id = existing.ID
	} else {
		id = uuid.New()
	}
	o := &domain.WorkshopOrder{
		ID:           id,
		ClientSlug:   slug,
		RequestedAt:  reqAt,
		DeliveryDate: delAt,
		Detail:       detail,
		TotalAmount:  total,
		IsBarter:     isBarter,
		Status:       st,
	}
	if existing != nil {
		o.DeliveredAt = existing.DeliveredAt
		o.CreatedAt = existing.CreatedAt
	}
	if st == domain.WorkshopEntregado && o.DeliveredAt == nil {
		now := time.Now()
		o.DeliveredAt = &now
	}
	return o, nil
}

func (s *Server) handleAdminPedidos(w http.ResponseWriter, r *http.Request) {
	wa, ok := s.workshopAdmin(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	list, err := wa.Orders.List(ctx)
	if err != nil {
		log.Error().Err(err).Msg("workshop list")
		http.Error(w, "err", http.StatusInternalServerError)
		return
	}
	stock, _ := wa.Filament.StockByColor(ctx)
	data := map[string]any{
		"Orders":     list,
		"Statuses":   workshopStatuses(),
		"Stock":      stock,
		"AdminToken": s.readAdminToken(r),
		"FlashError": strings.TrimSpace(r.URL.Query().Get("err")),
	}
	s.render(w, "admin_pedidos.html", data)
}

func (s *Server) handleAdminPedidosNew(w http.ResponseWriter, r *http.Request) {
	_, ok := s.workshopAdmin(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	s.renderAdminPedidosForm(w, r, false, nil, nil, "", 0)
}

func (s *Server) handleAdminPedidosCreate(w http.ResponseWriter, r *http.Request) {
	wa, ok := s.workshopAdmin(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	fil := parseFilamentsFromForm(r)
	o, err := workshopOrderFromForm(r, nil)
	if err != nil {
		s.renderAdminPedidosForm(w, r, false, nil, fil, err.Error(), 0)
		return
	}
	if err := wa.Orders.SaveOrderWithFilaments(ctx, o, fil); err != nil {
		if errors.Is(err, domain.ErrFilamentInsufficientStock) {
			s.renderAdminPedidosForm(w, r, false, o, fil, "Stock de filamento insuficiente para los gramos indicados.", 0)
			return
		}
		log.Error().Err(err).Msg("workshop create")
		http.Error(w, "err", http.StatusInternalServerError)
		return
	}
	if o.Status == domain.WorkshopEnImpresion {
		okStock, msg := s.validateStockForInPrint(ctx, nil, fil)
		if !okStock {
			_ = wa.Orders.UpdateStatus(ctx, o.ID, domain.WorkshopPendiente)
			o.Status = domain.WorkshopPendiente
			s.renderAdminPedidosForm(w, r, false, o, fil, msg+" El pedido se creó en estado pendiente.", 0)
			return
		}
	}
	http.Redirect(w, r, "/admin/pedidos", http.StatusFound)
}

func (s *Server) handleAdminPedidosEdit(w http.ResponseWriter, r *http.Request) {
	wa, ok := s.workshopAdmin(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	id, err := uuid.Parse(strings.TrimSpace(r.URL.Query().Get("id")))
	if err != nil {
		http.Error(w, "id", http.StatusBadRequest)
		return
	}
	o, err := wa.Orders.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "err", http.StatusInternalServerError)
		return
	}
	s.renderAdminPedidosForm(w, r, true, o, o.Filaments, "", 0)
}

func (s *Server) handleAdminPedidosSave(w http.ResponseWriter, r *http.Request) {
	wa, ok := s.workshopAdmin(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	id, err := uuid.Parse(strings.TrimSpace(r.FormValue("id")))
	if err != nil {
		http.Error(w, "id", http.StatusBadRequest)
		return
	}
	existing, err := wa.Orders.FindByID(ctx, id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	fil := parseFilamentsFromForm(r)
	o, err := workshopOrderFromForm(r, existing)
	if err != nil {
		o = existing
		if o != nil {
			o.Filaments = fil
		}
		s.renderAdminPedidosForm(w, r, true, o, fil, err.Error(), 0)
		return
	}
	o.ID = id
	o.Deposits = existing.Deposits
	if err := wa.Orders.SaveOrderWithFilaments(ctx, o, fil); err != nil {
		if errors.Is(err, domain.ErrFilamentInsufficientStock) {
			s.renderAdminPedidosForm(w, r, true, o, fil, "Stock de filamento insuficiente para los gramos indicados.", 0)
			return
		}
		log.Error().Err(err).Msg("workshop save")
		http.Error(w, "err", http.StatusInternalServerError)
		return
	}
	if existing.Status != domain.WorkshopEnImpresion && o.Status == domain.WorkshopEnImpresion {
		okStock, msg := s.validateStockForInPrint(ctx, existing.Filaments, fil)
		if !okStock {
			_ = wa.Orders.UpdateStatus(ctx, o.ID, existing.Status)
			o.Status = existing.Status
			s.renderAdminPedidosForm(w, r, true, o, fil, msg+" Se mantuvo el estado anterior.", 0)
			return
		}
	}
	http.Redirect(w, r, "/admin/pedidos", http.StatusFound)
}

func (s *Server) handleAdminPedidosDelete(w http.ResponseWriter, r *http.Request) {
	wa, ok := s.workshopAdmin(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form", http.StatusBadRequest)
		return
	}
	id, err := uuid.Parse(strings.TrimSpace(r.FormValue("id")))
	if err != nil {
		http.Error(w, "id", http.StatusBadRequest)
		return
	}
	if err := wa.Orders.Delete(r.Context(), id); err != nil {
		log.Error().Err(err).Msg("workshop delete")
		http.Error(w, "err", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/pedidos", http.StatusFound)
}

func (s *Server) handleAdminPedidosDeposit(w http.ResponseWriter, r *http.Request) {
	wa, ok := s.workshopAdmin(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	oid, err := uuid.Parse(strings.TrimSpace(r.FormValue("order_id")))
	if err != nil {
		http.Error(w, "order_id", http.StatusBadRequest)
		return
	}
	o, err := wa.Orders.FindByID(ctx, oid)
	if err != nil {
		http.Error(w, "pedido no encontrado", http.StatusNotFound)
		return
	}
	amt, err := parseMoneyAR(r.FormValue("amount"))
	if err != nil || amt <= 0 {
		http.Error(w, "monto inválido", http.StatusBadRequest)
		return
	}
	paidAt, err := parseWorkshopDate(r.FormValue("paid_at"))
	if err != nil {
		http.Error(w, "fecha seña inválida", http.StatusBadRequest)
		return
	}
	if !o.IsBarter && o.TotalAmount != nil {
		sum, _ := wa.Orders.SumDeposits(ctx, oid)
		if sum+amt > *o.TotalAmount+1e-6 {
			http.Error(w, "la suma de señas no puede superar el total del pedido", http.StatusBadRequest)
			return
		}
	}
	d := &domain.WorkshopDeposit{
		ID:              uuid.New(),
		WorkshopOrderID: oid,
		Amount:          amt,
		PaidAt:          paidAt,
	}
	if err := wa.Orders.AddDeposit(ctx, d); err != nil {
		log.Error().Err(err).Msg("deposit")
		http.Error(w, "err", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/pedidos/editar?id="+oid.String(), http.StatusFound)
}

func (s *Server) handleAdminPedidosStatus(w http.ResponseWriter, r *http.Request) {
	wa, ok := s.workshopAdmin(w, r)
	if !ok {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form", http.StatusBadRequest)
		return
	}
	id, err := uuid.Parse(strings.TrimSpace(r.FormValue("id")))
	if err != nil {
		http.Error(w, "id", http.StatusBadRequest)
		return
	}
	st := domain.WorkshopOrderStatus(strings.TrimSpace(r.FormValue("status")))
	valid := false
	for _, v := range workshopStatuses() {
		if st == v {
			valid = true
			break
		}
	}
	if !valid {
		http.Error(w, "estado inválido", http.StatusBadRequest)
		return
	}
	if st == domain.WorkshopEnImpresion {
		o, err := wa.Orders.FindByID(r.Context(), id)
		if err != nil {
			http.Error(w, "pedido no encontrado", http.StatusNotFound)
			return
		}
		okStock, msg := s.validateOrderCanEnterInPrint(r.Context(), o)
		if !okStock {
			http.Redirect(w, r, "/admin/pedidos?err="+url.QueryEscape(msg), http.StatusFound)
			return
		}
	}
	if err := wa.Orders.UpdateStatus(r.Context(), id, st); err != nil {
		log.Error().Err(err).Msg("workshop status")
		http.Error(w, "err", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/pedidos", http.StatusFound)
}

// --- Telegram webhook ---

type tgWebhookMsg struct {
	Message *struct {
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
		Text string `json:"text"`
	} `json:"message"`
}

func telegramAdminChatIDs() map[string]struct{} {
	out := map[string]struct{}{}
	raw := os.Getenv("TELEGRAM_CHAT_IDS")
	if strings.TrimSpace(raw) == "" {
		raw = os.Getenv("TELEGRAM_CHAT_ID")
	}
	for _, part := range strings.Split(raw, ",") {
		id := strings.TrimSpace(part)
		if id != "" {
			out[id] = struct{}{}
		}
	}
	return out
}

func isTelegramAdminChat(chatID int64) bool {
	key := strconv.FormatInt(chatID, 10)
	_, ok := telegramAdminChatIDs()[key]
	return ok
}

func parseWorkshopStatusToken(tok string) (domain.WorkshopOrderStatus, bool) {
	t := strings.TrimSpace(strings.ToLower(strings.Trim(tok, "'\"`")))
	aliases := map[string]domain.WorkshopOrderStatus{
		"pendiente":       domain.WorkshopPendiente,
		"disenado":        domain.WorkshopDisenado,
		"diseñado":        domain.WorkshopDisenado,
		"diseno":          domain.WorkshopDisenado,
		"en_impresion":    domain.WorkshopEnImpresion,
		"impresion":       domain.WorkshopEnImpresion,
		"listo_entrega":   domain.WorkshopListoEntrega,
		"listo":           domain.WorkshopListoEntrega,
		"entregado":       domain.WorkshopEntregado,
	}
	st, ok := aliases[t]
	return st, ok
}

func (s *Server) handleTelegramWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method", http.StatusMethodNotAllowed)
		return
	}
	secret := strings.TrimSpace(os.Getenv("TELEGRAM_WEBHOOK_SECRET"))
	if secret != "" {
		got := strings.TrimSpace(r.Header.Get("X-Telegram-Bot-Api-Secret-Token"))
		if subtle.ConstantTimeCompare([]byte(got), []byte(secret)) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 65536))
	if err != nil {
		http.Error(w, "body", http.StatusBadRequest)
		return
	}
	var up tgWebhookMsg
	if err := json.Unmarshal(body, &up); err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	if up.Message == nil || up.Message.Text == "" {
		w.WriteHeader(http.StatusOK)
		return
	}
	text := strings.TrimSpace(up.Message.Text)
	parts := strings.Fields(text)
	if len(parts) == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}
	cmd := parts[0]
	if i := strings.Index(cmd, "@"); i > 0 {
		cmd = cmd[:i]
	}
	if cmd != "/estado" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if s.workshop == nil {
		_ = telegram.SendToChat(strconv.FormatInt(up.Message.Chat.ID, 10), "módulo pedidos no disponible")
		w.WriteHeader(http.StatusOK)
		return
	}
	if !isTelegramAdminChat(up.Message.Chat.ID) {
		_ = telegram.SendToChat(strconv.FormatInt(up.Message.Chat.ID, 10), "no autorizado")
		w.WriteHeader(http.StatusOK)
		return
	}
	if len(parts) < 3 {
		_ = telegram.SendToChat(strconv.FormatInt(up.Message.Chat.ID, 10), "uso: /estado <estado> <cliente_snake_case>\nej: /estado en_impresion fede_bertoqui")
		w.WriteHeader(http.StatusOK)
		return
	}
	st, ok := parseWorkshopStatusToken(parts[1])
	if !ok {
		_ = telegram.SendToChat(strconv.FormatInt(up.Message.Chat.ID, 10), "estado no reconocido. Usá: pendiente, disenado, en_impresion, listo_entrega, entregado")
		w.WriteHeader(http.StatusOK)
		return
	}
	slug, err := normalizeWorkshopClientSlug(parts[2])
	if err != nil {
		_ = telegram.SendToChat(strconv.FormatInt(up.Message.Chat.ID, 10), err.Error())
		w.WriteHeader(http.StatusOK)
		return
	}
	ctx := context.Background()
	list, err := s.workshop.Orders.FindUndeliveredByClientSlug(ctx, slug)
	if err != nil || len(list) == 0 {
		_ = telegram.SendToChat(strconv.FormatInt(up.Message.Chat.ID, 10), "no hay pedidos activos para ese cliente")
		w.WriteHeader(http.StatusOK)
		return
	}
	if len(list) > 1 {
		var b strings.Builder
		b.WriteString("varios pedidos activos; especificá id en el admin:\n")
		for _, o := range list {
			fmt.Fprintf(&b, "- %s estado=%s entrega=%s\n", o.ID.String()[:8], o.Status, o.DeliveryDate.Format("2006-01-02"))
		}
		_ = telegram.SendToChat(strconv.FormatInt(up.Message.Chat.ID, 10), b.String())
		w.WriteHeader(http.StatusOK)
		return
	}
	if err := s.workshop.Orders.UpdateStatus(ctx, list[0].ID, st); err != nil {
		_ = telegram.SendToChat(strconv.FormatInt(up.Message.Chat.ID, 10), "error actualizando")
		w.WriteHeader(http.StatusOK)
		return
	}
	_ = telegram.SendToChat(strconv.FormatInt(up.Message.Chat.ID, 10), fmt.Sprintf("ok: %s -> %s", slug, st))
	w.WriteHeader(http.StatusOK)
}
