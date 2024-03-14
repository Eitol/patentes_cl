package patentes_cl

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

type Vehicle struct {
	Ppu             string `json:"ppu"`
	Marca           string `json:"marca"`
	Modelo          string `json:"modelo"`
	Tipo            string `json:"tipo"`
	AFabricacion    string `json:"aFabricacion"`
	NroMotor        string `json:"nroMotor"`
	NroChasis       string `json:"nroChasis"`
	NroSerie        string `json:"nroSerie"`
	NroVin          string `json:"nroVin"`
	CodigoColorBase string `json:"codigoColorBase"`
	DescColorBase   string `json:"descColorBase"`
	RestoColor      string `json:"restoColor"`
	Calidad         string `json:"calidad"`
	DvPpu           string `json:"dvPpu"`
	TipoPropietario string `json:"tipoPropietario"`
}
