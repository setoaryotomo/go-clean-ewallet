package models

type ResponseFindPedagangKiosGradingWeekly struct {
	KodeGrading   string `json:"kodegrading"`
	NamaCorporate string `json:"namacorporate"`
	NamaPedagang  string `json:"namapedagang"`
	NoReg         string `json:"noreg"`
	Year          int    `json:"year"`
	Week          int    `json:"week"`
	Poin          int    `json:"poin"`
	Grade         string `json:"grade"`
}

type WeekPoinBonus struct {
	IDCorporate int `json:"idcorporate"`
	Week        int `json:"week"`
	BonusPoin   int `json:"bonuspoin"`
}

type PedagangKiosGradingWeekly struct {
	IDPedagangKios        int    `json:"idpedagangkios"`
	NamaPedagang          string `json:"namapedagang"`
	NoReg                 string `json:"noreg"`
	Week                  int    `json:"week"`
	Poin                  int    `json:"poin"`
	IDTagihanKategori     int    `json:"idtagihankategori"`
	UraianTagihanKategori string `json:"uraiantagihankategori"`
	BulanMingguIni        int    `json:"bulanmingguini"`
	BulanMingguDepan      int    `json:"bulanminggudepan"`
	IDCorporate           int    `json:"idcorporate"`
	CID                   string `json:"cid"`
	NamaCorporate         string `json:"namacorporate"`
	Year                  int    `json:"year"`
}
