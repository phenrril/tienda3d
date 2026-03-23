package httpserver

import "github.com/phenrril/tienda3d/internal/domain"

// WorkshopAdmin agrupa repositorios para pedidos de taller, filamento y gastos.
type WorkshopAdmin struct {
	Orders   domain.WorkshopRepo
	Filament domain.FilamentLedgerRepo
	Expenses domain.BusinessExpenseRepo
	Settings domain.AppSettingRepo
}
