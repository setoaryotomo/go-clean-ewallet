package pedagangKiosPoinRepository

import (
	"database/sql"
	"sample/helpers"
	"sample/models"
	"sample/repositories"
	"strings"
	"time"
)

type pedagangKiosPoinRepository struct {
	RepoDB repositories.Repository
}

func NewPedagangKiosPoinRepository(RepoDB repositories.Repository) pedagangKiosPoinRepository {
	return pedagangKiosPoinRepository{
		RepoDB: RepoDB,
	}
}

var defineColumn = `id, idpedagangkios, namapedagang, noreg, idtagihankategori, uraiantagihankategori,
	tanggaltransaksi, poin, idcorporate, cid, namacorporate, pedagangkiospoinunique, created_at`

// FindPedagangKiosDataBulkByDate
func (ctx pedagangKiosPoinRepository) FindPedagangKiosDataByDate(date string) ([]models.PedagangKiosPoin, error) {
	var result []models.PedagangKiosPoin
	datestart := date + " 00:00:00"
	dateend := date + " 23:59:59"

	var query = `
	SELECT tgp.idpedagangkios, pdg.namapedagang, k.noreg, tgp.idtagihankategori,
		tgk.uraian, td.created_at::date::varchar, c.id, c.cid, c.uraian
	FROM trxdetail td
		JOIN trx t ON td.idtrx = t.id
		JOIN tagihanpedagang tgp ON td.idtagihanpedagang = tgp.id
		JOIN pedagangkios pdk ON t.idpedagangkios = pdk.id
		JOIN kios k ON pdk.idkios = k.id
		JOIN pedagang pdg ON pdk.idpedagang = pdg.id
		JOIN tagihankategori tgk ON tgp.idtagihankategori = tgk.id
		JOIN corporate c ON t.idcorporate = c.id
	WHERE td.deleted_at IS NULL AND td.created_at BETWEEN $1 AND $2 AND t.statustrx = 1
	GROUP BY tgp.idpedagangkios, tgp.idtagihankategori, pdg.namapedagang,
		k.noreg, td.created_at::date, tgk.uraian, c.id, c.cid, c.uraian
	ORDER BY td.created_at::date, tgp.idpedagangkios ASC`

	rows, err := ctx.RepoDB.DB.Query(query, datestart, dateend)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var val models.PedagangKiosPoin
		err := rows.Scan(&val.IDPedagangKios, &val.NamaPedagang, &val.NoReg, &val.IDTagihanKategori,
			&val.UraianTagihanKategori, &val.TanggalTransaksi, &val.IDCorporate, &val.CID, &val.NamaCorporate)
		if err != nil {
			return result, err
		}
		result = append(result, val)
	}

	return result, nil
}

// IsPedagangKiosPoinExistsByIndex
func (ctx pedagangKiosPoinRepository) IsPedagangKiosPoinExistsByIndex(unique string) (models.PedagangKiosPoin, error) {
	var result models.PedagangKiosPoin
	var query = `SELECT ` + defineColumn + ` FROM pedagangkiospoin WHERE pedagangkiospoinunique = $1`
	rows, err := ctx.RepoDB.DB.Query(query, unique)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	err = rows.Err()
	if err != nil {
		return result, err
	}
	data, err := pedagangKiosPoinDto(rows)
	if len(data) == 0 {
		return result, err
	}
	if err != nil {
		return result, err
	}

	return data[0], nil
}

// AddPedagangKiosPoin
func (ctx pedagangKiosPoinRepository) AddPedagangKiosPoin(req []models.PedagangKiosPoin) (bool, error) {
	var result bool

	sqlStr := `INSERT INTO pedagangkiospoin (created_at, idpedagangkios, namapedagang, noreg,
		idtagihankategori, uraiantagihankategori, tanggaltransaksi, poin, pedagangkiospoinunique,
		idcorporate, cid, namacorporate) VALUES `
	vals := []interface{}{}
	t := time.Now()
	dbTime := t.Format("2006-01-02 15:04:05")
	for _, row := range req {
		sqlStr += "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?),"

		vals = append(vals, dbTime, row.IDPedagangKios, row.NamaPedagang, row.NoReg, row.IDTagihanKategori,
			row.UraianTagihanKategori, row.TanggalTransaksi, row.Poin, row.PedagangKiosPoinUnique, row.IDCorporate,
			row.CID, row.NamaCorporate)
	}
	//trim the last ,
	sqlStr = strings.TrimSuffix(sqlStr, ",")

	//Replacing ? with $n for postgres
	sqlStr = helpers.ReplaceSQL(sqlStr, "?")
	//prepare the statement
	stmt, _ := ctx.RepoDB.DB.Prepare(sqlStr)
	//format all vals at once
	rows, err := stmt.Query(vals...)
	if err != nil {
		// handle this error
		result = false
		return result, err
	}
	defer rows.Close()

	result = true

	return result, nil
}

// pedagangKiosPoinDto
func pedagangKiosPoinDto(rows *sql.Rows) ([]models.PedagangKiosPoin, error) {
	var result []models.PedagangKiosPoin

	for rows.Next() {
		var val models.PedagangKiosPoin
		err := rows.Scan(&val.ID, &val.IDPedagangKios, &val.NamaPedagang, &val.NoReg, &val.IDTagihanKategori, &val.UraianTagihanKategori,
			&val.TanggalTransaksi, &val.Poin, &val.PedagangKiosPoinUnique, &val.Created_at, &val.IDPedagangKios, &val.Updated_at)
		if err != nil {
			return result, err
		}
		result = append(result, val)
	}

	return result, nil
}
