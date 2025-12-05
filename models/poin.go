package models

type RequestAddPedagangKiosPoin struct {
	IDPedagangKios         int    `json:"idpedagangkios"  validate:"required"`
	NamaPedagang           string `json:"namapedagang"  validate:"required"`
	NoReg                  string `json:"noreg"  validate:"required"`
	IDTagihanKategori      int    `json:"idtagihankategori"  validate:"required"`
	UraianTagihanKategori  string `json:"uraiantagihankategori"  validate:"required"`
	TanggalTransaksi       string `json:"tanggaltransaksi"  validate:"required"`
	Poin                   int    `json:"poin" validate:"required"`
	PedagangKiosPoinUnique string `json:"unique" validate:"required"`
	IDCorporate            int    `json:"idcorporate" validate:"required"`
	CID                    string `json:"cid" validate:"required"`
	NamaCorporate          string `json:"namacorporate" validate:"required"`
}

type PedagangKiosPoinData struct {
	IDPedagangKios        int    `json:"idpedagangkios"`
	NamaPedagang          string `json:"namapedagang"`
	NoReg                 string `json:"noreg"`
	IDTagihanKategori     int    `json:"idtagihankategori"`
	UraianTagihanKategori string `json:"uraiantagihankategori"`
	TanggalTransaksi      string `json:"tanggaltransaksi"`
	IDCorporate           int    `json:"idcorporate"`
	CID                   string `json:"cid"`
	NamaCorporate         string `json:"namacorporate"`
}

type PedagangKiosPoin struct {
	ID                     int       `json:"id"`
	IDPedagangKios         int       `json:"idpedagangkios"`
	NamaPedagang           string    `json:"namapedagang"`
	NoReg                  string    `json:"noreg"`
	IDTagihanKategori      int       `json:"idtagihankategori"`
	UraianTagihanKategori  string    `json:"uraiantagihankategori"`
	TanggalTransaksi       string    `json:"tanggaltransaksi"`
	Poin                   int       `json:"poin"`
	IDCorporate            int       `json:"idcorporate"`
	CID                    string    `json:"cid"`
	NamaCorporate          string    `json:"namacorporate"`
	PedagangKiosPoinUnique string    `json:"pedagangkiospoinunique"`
	Created_at             JSONTime  `json:"created_at"`
	Updated_at             *JSONTime `json:"updated_at"`
}
