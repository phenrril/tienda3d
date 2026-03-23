package domain

import "errors"

var ErrNotFound = errors.New("not found")

// ErrFilamentInsufficientStock indica que no hay gramos suficientes en inventario para el consumo pedido.
var ErrFilamentInsufficientStock = errors.New("filamento: stock insuficiente")
