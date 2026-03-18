package analytics

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	analyticsdata "google.golang.org/api/analyticsdata/v1beta"
	"google.golang.org/api/option"
)

type DailyRow struct {
	Date           string
	ActiveUsers    int64
	AvgSessionSec  float64
	WhatsAppClicks int64
}

type CategoryRow struct {
	Name  string
	Count int64
}

type TrafficRow struct {
	Source   string
	Medium  string
	Campaign string
	Sessions int64
}

type DashboardData struct {
	From string
	To   string

	TotalUsers          int64
	TotalWhatsAppClicks int64
	AvgSessionDuration  float64

	Daily      []DailyRow
	Categories []CategoryRow
	Traffic    []TrafficRow

	Error    string
	Warnings []string
}

type Client struct {
	svc        *analyticsdata.Service
	propertyID string
}

func NewClient() *Client {
	propertyID := strings.TrimSpace(os.Getenv("GA4_PROPERTY_ID"))
	credsFile := strings.TrimSpace(os.Getenv("GOOGLE_ANALYTICS_CREDENTIALS"))

	if propertyID == "" {
		log.Warn().Msg("GA4_PROPERTY_ID no configurado — analytics dashboard deshabilitado")
		return &Client{}
	}

	var opts []option.ClientOption
	if credsFile != "" {
		opts = append(opts, option.WithCredentialsFile(credsFile))
	}

	svc, err := analyticsdata.NewService(context.Background(), opts...)
	if err != nil {
		log.Error().Err(err).Msg("no se pudo crear cliente GA4 Data API")
		return &Client{}
	}

	log.Info().Str("property", propertyID).Msg("GA4 Data API habilitada")
	return &Client{svc: svc, propertyID: propertyID}
}

func (c *Client) Available() bool {
	return c.svc != nil && c.propertyID != ""
}

func (c *Client) FetchDashboard(ctx context.Context, from, to time.Time) DashboardData {
	d := DashboardData{
		From: from.Format("2006-01-02"),
		To:   to.Format("2006-01-02"),
	}

	if !c.Available() {
		d.Error = "GA4 no configurado. Configurá GA4_PROPERTY_ID y GOOGLE_ANALYTICS_CREDENTIALS."
		return d
	}

	prop := fmt.Sprintf("properties/%s", c.propertyID)
	dateRange := &analyticsdata.DateRange{
		StartDate: from.Format("2006-01-02"),
		EndDate:   to.Format("2006-01-02"),
	}

	c.fetchDailyUsers(ctx, prop, dateRange, &d)
	c.fetchWhatsAppDaily(ctx, prop, dateRange, &d)
	c.fetchCategories(ctx, prop, dateRange, &d)
	c.fetchTraffic(ctx, prop, dateRange, &d)

	return d
}

func (c *Client) fetchDailyUsers(ctx context.Context, prop string, dr *analyticsdata.DateRange, d *DashboardData) {
	req := &analyticsdata.RunReportRequest{
		DateRanges: []*analyticsdata.DateRange{dr},
		Dimensions: []*analyticsdata.Dimension{
			{Name: "date"},
		},
		Metrics: []*analyticsdata.Metric{
			{Name: "activeUsers"},
			{Name: "averageSessionDuration"},
		},
		OrderBys: []*analyticsdata.OrderBy{
			{Dimension: &analyticsdata.DimensionOrderBy{DimensionName: "date"}},
		},
	}

	resp, err := c.svc.Properties.RunReport(prop, req).Context(ctx).Do()
	if err != nil {
		log.Error().Err(err).Msg("GA4: fetchDailyUsers")
		d.Error = "Error consultando usuarios diarios: " + err.Error()
		return
	}

	var totalUsers int64
	var totalDuration float64
	var daysWithDuration int

	for _, row := range resp.Rows {
		date := row.DimensionValues[0].Value
		users := parseInt64(row.MetricValues[0].Value)
		avgSec := parseFloat64(row.MetricValues[1].Value)

		totalUsers += users
		if avgSec > 0 {
			totalDuration += avgSec
			daysWithDuration++
		}

		existing := findOrAppendDaily(d, date)
		existing.ActiveUsers = users
		existing.AvgSessionSec = avgSec
	}

	d.TotalUsers = totalUsers
	if daysWithDuration > 0 {
		d.AvgSessionDuration = totalDuration / float64(daysWithDuration)
	}
}

func (c *Client) fetchWhatsAppDaily(ctx context.Context, prop string, dr *analyticsdata.DateRange, d *DashboardData) {
	req := &analyticsdata.RunReportRequest{
		DateRanges: []*analyticsdata.DateRange{dr},
		Dimensions: []*analyticsdata.Dimension{
			{Name: "date"},
		},
		Metrics: []*analyticsdata.Metric{
			{Name: "eventCount"},
		},
		DimensionFilter: &analyticsdata.FilterExpression{
			Filter: &analyticsdata.Filter{
				FieldName: "eventName",
				StringFilter: &analyticsdata.StringFilter{
					MatchType: "EXACT",
					Value:     "whatsapp_click",
				},
			},
		},
		OrderBys: []*analyticsdata.OrderBy{
			{Dimension: &analyticsdata.DimensionOrderBy{DimensionName: "date"}},
		},
	}

	resp, err := c.svc.Properties.RunReport(prop, req).Context(ctx).Do()
	if err != nil {
		log.Error().Err(err).Msg("GA4: fetchWhatsAppDaily")
		return
	}

	var total int64
	for _, row := range resp.Rows {
		date := row.DimensionValues[0].Value
		count := parseInt64(row.MetricValues[0].Value)
		total += count

		existing := findOrAppendDaily(d, date)
		existing.WhatsAppClicks = count
	}
	d.TotalWhatsAppClicks = total
}

func (c *Client) fetchCategories(ctx context.Context, prop string, dr *analyticsdata.DateRange, d *DashboardData) {
	// Intentamos primero con el parámetro de evento personalizado.
	// Requiere que "category_name" esté registrado como dimensión personalizada en GA4
	// (Admin → Definiciones personalizadas → Crear dimensión personalizada → Ámbito: Evento → Nombre del parámetro: category_name)
	// o que haya eventos recientes (últimos 60 días) con ese parámetro.
	req := &analyticsdata.RunReportRequest{
		DateRanges: []*analyticsdata.DateRange{dr},
		Dimensions: []*analyticsdata.Dimension{
			{Name: "customEvent:category_name"},
		},
		Metrics: []*analyticsdata.Metric{
			{Name: "eventCount"},
		},
		DimensionFilter: &analyticsdata.FilterExpression{
			Filter: &analyticsdata.Filter{
				FieldName: "eventName",
				StringFilter: &analyticsdata.StringFilter{
					MatchType: "EXACT",
					Value:     "category_click",
				},
			},
		},
		OrderBys: []*analyticsdata.OrderBy{
			{Metric: &analyticsdata.MetricOrderBy{MetricName: "eventCount"}, Desc: true},
		},
		Limit: 20,
	}

	resp, err := c.svc.Properties.RunReport(prop, req).Context(ctx).Do()
	if err != nil {
		log.Error().Err(err).Msg("GA4: fetchCategories con customEvent:category_name")
		d.Warnings = append(d.Warnings, "Categorías: "+err.Error())
		return
	}

	for _, row := range resp.Rows {
		name := row.DimensionValues[0].Value
		if name == "" || name == "(not set)" {
			continue
		}
		d.Categories = append(d.Categories, CategoryRow{
			Name:  name,
			Count: parseInt64(row.MetricValues[0].Value),
		})
	}

	// Si no hay filas útiles, registramos aviso para ayudar al diagnóstico.
	if len(d.Categories) == 0 {
		d.Warnings = append(d.Warnings, `Categorías: no hay datos. Si nunca se registraron eventos "category_click" en este período, o el parámetro "category_name" no está registrado como dimensión personalizada en GA4, este panel estará vacío.`)
	}
}

func (c *Client) fetchTraffic(ctx context.Context, prop string, dr *analyticsdata.DateRange, d *DashboardData) {
	req := &analyticsdata.RunReportRequest{
		DateRanges: []*analyticsdata.DateRange{dr},
		Dimensions: []*analyticsdata.Dimension{
			{Name: "sessionSource"},
			{Name: "sessionMedium"},
			{Name: "sessionCampaignName"},
		},
		Metrics: []*analyticsdata.Metric{
			{Name: "sessions"},
		},
		OrderBys: []*analyticsdata.OrderBy{
			{Metric: &analyticsdata.MetricOrderBy{MetricName: "sessions"}, Desc: true},
		},
		Limit: 30,
	}

	resp, err := c.svc.Properties.RunReport(prop, req).Context(ctx).Do()
	if err != nil {
		log.Error().Err(err).Msg("GA4: fetchTraffic")
		d.Warnings = append(d.Warnings, "Tráfico: "+err.Error())
		return
	}

	for _, row := range resp.Rows {
		d.Traffic = append(d.Traffic, TrafficRow{
			Source:   row.DimensionValues[0].Value,
			Medium:   row.DimensionValues[1].Value,
			Campaign: row.DimensionValues[2].Value,
			Sessions: parseInt64(row.MetricValues[0].Value),
		})
	}

	if len(d.Traffic) == 0 {
		d.Warnings = append(d.Warnings, "Tráfico: la consulta no devolvió filas para este período.")
	}
}

func findOrAppendDaily(d *DashboardData, date string) *DailyRow {
	for i := range d.Daily {
		if d.Daily[i].Date == date {
			return &d.Daily[i]
		}
	}
	d.Daily = append(d.Daily, DailyRow{Date: date})
	return &d.Daily[len(d.Daily)-1]
}

func parseInt64(s string) int64 {
	var n int64
	fmt.Sscanf(s, "%d", &n)
	return n
}

func parseFloat64(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
