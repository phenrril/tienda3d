package smtp

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"os"
	"strings"

	"gopkg.in/gomail.v2"

	"github.com/phenrril/tienda3d/internal/domain"
)

type SMTPService struct {
	host     string
	port     int
	user     string
	password string
	from     string
}

func NewSMTPService() *SMTPService {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		host = "smtp.gmail.com"
	}
	
	port := 587
	portStr := os.Getenv("SMTP_PORT")
	if portStr != "" {
		fmt.Sscanf(portStr, "%d", &port)
	}

	user := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")
	if from == "" {
		from = user
	}

	return &SMTPService{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		from:     from,
	}
}

func (s *SMTPService) SendOrderConfirmation(ctx context.Context, order *domain.Order) error {
	if s.user == "" || s.password == "" {
		// No hay configuración SMTP, simplemente loguear y no fallar
		fmt.Printf("⚠️  SMTP no configurado - no se envió email para orden %s\n", order.ID)
		return nil
	}

	// Generar HTML del email
	htmlBody, err := s.generateOrderHTML(order)
	if err != nil {
		return fmt.Errorf("error generando HTML del email: %w", err)
	}

	// Crear mensaje
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", order.Email)
	m.SetHeader("Subject", fmt.Sprintf("✅ Confirmación de tu pedido #%s", order.ID.String()[:8]))
	m.SetBody("text/html", htmlBody)

	// Enviar
	d := gomail.NewDialer(s.host, s.port, s.user, s.password)
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("error enviando email: %w", err)
	}

	fmt.Printf("📧 Email de confirmación enviado a %s para orden %s\n", order.Email, order.ID)
	return nil
}

func (s *SMTPService) generateOrderHTML(order *domain.Order) (string, error) {
	tmplStr := `
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Confirmación de Pedido</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; background-color: #f3f4f6;">
    <table role="presentation" style="width: 100%; border-collapse: collapse;">
        <tr>
            <td style="padding: 40px 20px; text-align: center;">
                <table role="presentation" style="max-width: 600px; margin: 0 auto; background-color: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);">
                    <!-- Header -->
                    <tr>
                        <td style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 40px 30px; text-align: center;">
                            <h1 style="margin: 0; color: #ffffff; font-size: 28px; font-weight: bold;">¡Gracias por tu compra!</h1>
                            <p style="margin: 10px 0 0 0; color: #e0e7ff; font-size: 16px;">Tu pedido ha sido confirmado</p>
                        </td>
                    </tr>
                    
                    <!-- Contenido -->
                    <tr>
                        <td style="padding: 40px 30px;">
                            <p style="margin: 0 0 20px 0; color: #374151; font-size: 16px; line-height: 1.6;">
                                Hola <strong>{{.Name}}</strong>,
                            </p>
                            <p style="margin: 0 0 30px 0; color: #374151; font-size: 16px; line-height: 1.6;">
                                Hemos recibido tu pedido correctamente. A continuación encontrarás el detalle de tu compra:
                            </p>
                            
                            <!-- Información del Pedido -->
                            <table role="presentation" style="width: 100%; border-collapse: collapse; margin-bottom: 30px; background-color: #f9fafb; border-radius: 6px; overflow: hidden;">
                                <tr>
                                    <td style="padding: 20px; border-bottom: 1px solid #e5e7eb;">
                                        <p style="margin: 0; color: #6b7280; font-size: 14px;">Número de Pedido</p>
                                        <p style="margin: 5px 0 0 0; color: #111827; font-size: 18px; font-weight: bold;">#{{.OrderNumber}}</p>
                                    </td>
                                </tr>
                                <tr>
                                    <td style="padding: 20px;">
                                        <p style="margin: 0; color: #6b7280; font-size: 14px;">Método de Pago</p>
                                        <p style="margin: 5px 0 0 0; color: #111827; font-size: 16px;">{{.PaymentMethod}}</p>
                                    </td>
                                </tr>
                            </table>
                            
                            <!-- Items del Pedido -->
                            <h2 style="margin: 0 0 20px 0; color: #111827; font-size: 20px; font-weight: bold;">Detalle de Productos</h2>
                            <table role="presentation" style="width: 100%; border-collapse: collapse; margin-bottom: 30px;">
                                <thead>
                                    <tr style="border-bottom: 2px solid #e5e7eb;">
                                        <th style="padding: 12px; text-align: left; color: #6b7280; font-size: 14px; font-weight: 600;">Producto</th>
                                        <th style="padding: 12px; text-align: center; color: #6b7280; font-size: 14px; font-weight: 600;">Cant.</th>
                                        <th style="padding: 12px; text-align: right; color: #6b7280; font-size: 14px; font-weight: 600;">Precio</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {{range .Items}}
                                    <tr style="border-bottom: 1px solid #f3f4f6;">
                                        <td style="padding: 16px 12px;">
                                            <p style="margin: 0; color: #111827; font-size: 15px; font-weight: 500;">{{.Title}}</p>
                                            {{if .Color}}
                                            <p style="margin: 4px 0 0 0; color: #6b7280; font-size: 13px;">Color: {{.Color}}</p>
                                            {{end}}
                                        </td>
                                        <td style="padding: 16px 12px; text-align: center; color: #374151; font-size: 15px;">{{.Qty}}</td>
                                        <td style="padding: 16px 12px; text-align: right; color: #111827; font-size: 15px; font-weight: 500;">${{printf "%.2f" .Subtotal}}</td>
                                    </tr>
                                    {{end}}
                                </tbody>
                            </table>
                            
                            <!-- Totales -->
                            <table role="presentation" style="width: 100%; border-collapse: collapse; margin-bottom: 30px;">
                                {{if .ShippingCost}}
                                <tr>
                                    <td style="padding: 8px 0; text-align: right; color: #6b7280; font-size: 15px;">Envío:</td>
                                    <td style="padding: 8px 0 8px 20px; text-align: right; color: #111827; font-size: 15px; font-weight: 500; width: 100px;">${{printf "%.2f" .ShippingCost}}</td>
                                </tr>
                                {{end}}
                                {{if .DiscountAmount}}
                                <tr>
                                    <td style="padding: 8px 0; text-align: right; color: #059669; font-size: 15px;">Descuento:</td>
                                    <td style="padding: 8px 0 8px 20px; text-align: right; color: #059669; font-size: 15px; font-weight: 500; width: 100px;">-${{printf "%.2f" .DiscountAmount}}</td>
                                </tr>
                                {{end}}
                                <tr style="border-top: 2px solid #e5e7eb;">
                                    <td style="padding: 16px 0; text-align: right; color: #111827; font-size: 18px; font-weight: bold;">Total:</td>
                                    <td style="padding: 16px 0 16px 20px; text-align: right; color: #667eea; font-size: 20px; font-weight: bold; width: 100px;">${{printf "%.2f" .Total}}</td>
                                </tr>
                            </table>
                            
                            <!-- Dirección de Envío -->
                            {{if .Address}}
                            <div style="background-color: #f9fafb; padding: 20px; border-radius: 6px; margin-bottom: 30px;">
                                <h3 style="margin: 0 0 12px 0; color: #111827; font-size: 16px; font-weight: bold;">Dirección de Envío</h3>
                                <p style="margin: 0; color: #374151; font-size: 14px; line-height: 1.6;">
                                    {{.Address}}<br>
                                    {{if .PostalCode}}CP: {{.PostalCode}}<br>{{end}}
                                    {{if .Province}}{{.Province}}{{end}}
                                </p>
                            </div>
                            {{end}}
                            
                            <!-- Próximos Pasos -->
                            <div style="background-color: #eff6ff; border-left: 4px solid #3b82f6; padding: 20px; border-radius: 4px; margin-bottom: 20px;">
                                <h3 style="margin: 0 0 12px 0; color: #1e40af; font-size: 16px; font-weight: bold;">📦 Próximos Pasos</h3>
                                <p style="margin: 0; color: #1e40af; font-size: 14px; line-height: 1.6;">
                                    Te contactaremos pronto para coordinar los detalles de tu pedido. 
                                    Si tenés alguna consulta, no dudes en contactarnos.
                                </p>
                            </div>
                        </td>
                    </tr>
                    
                    <!-- Footer -->
                    <tr>
                        <td style="background-color: #f9fafb; padding: 30px; text-align: center; border-top: 1px solid #e5e7eb;">
                            <p style="margin: 0 0 10px 0; color: #6b7280; font-size: 14px;">
                                Gracias por confiar en nosotros
                            </p>
                            <p style="margin: 0; color: #9ca3af; font-size: 12px;">
                                Este es un email automático, por favor no responder.
                            </p>
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`

	// Preparar datos para el template
	type ItemData struct {
		Title    string
		Color    string
		Qty      int
		Subtotal float64
	}

	items := make([]ItemData, len(order.Items))
	for i, item := range order.Items {
		items[i] = ItemData{
			Title:    item.Title,
			Color:    item.Color,
			Qty:      item.Qty,
			Subtotal: item.UnitPrice * float64(item.Qty),
		}
	}

	paymentMethodName := "No especificado"
	switch strings.ToLower(order.PaymentMethod) {
	case "mercadopago":
		paymentMethodName = "MercadoPago"
	case "efectivo":
		paymentMethodName = "Efectivo"
	case "transferencia", "transfer":
		paymentMethodName = "Transferencia Bancaria"
	default:
		if order.PaymentMethod != "" {
			paymentMethodName = order.PaymentMethod
		}
	}

	data := struct {
		Name           string
		OrderNumber    string
		PaymentMethod  string
		Items          []ItemData
		ShippingCost   float64
		DiscountAmount float64
		Total          float64
		Address        string
		PostalCode     string
		Province       string
	}{
		Name:           order.Name,
		OrderNumber:    order.ID.String()[:8],
		PaymentMethod:  paymentMethodName,
		Items:          items,
		ShippingCost:   order.ShippingCost,
		DiscountAmount: order.DiscountAmount,
		Total:          order.Total,
		Address:        order.Address,
		PostalCode:     order.PostalCode,
		Province:       order.Province,
	}

	tmpl, err := template.New("email").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

