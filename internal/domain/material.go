package domain

// Material enum soportado.
// Coeficientes usados por PricingService simple.
type Material string

const (
	MaterialPLA  Material = "PLA"
	MaterialPETG Material = "PETG"
	MaterialTPU  Material = "TPU"
)

// MaterialConf guarda coeficientes de precio.
type MaterialConf struct {
	Material   Material
	CoefPerCM3 float64
	CoefPerMin float64 // opcional si se diferencia por material
}
