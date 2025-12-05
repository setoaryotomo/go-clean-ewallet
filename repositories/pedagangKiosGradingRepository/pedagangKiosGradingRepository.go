package pedagangKiosGradingRepository

import (
	"sample/helpers"
	"sample/models"
	"sample/repositories"
	"strconv"
	"strings"
	"time"
)

type pedagangKiosGradingRepository struct {
	RepoDB repositories.Repository
}

func NewPedagangKiosGradingRepository(RepoDB repositories.Repository) pedagangKiosGradingRepository {
	return pedagangKiosGradingRepository{
		RepoDB: RepoDB,
	}
}

// FindPedagangKiosGradingWeekly
func (ctx pedagangKiosGradingRepository) FindPedagangKiosGradingWeekly(idcorporate, week int) ([]models.PedagangKiosGradingWeekly, error) {
	var result []models.PedagangKiosGradingWeekly
	stridcorporate := strconv.Itoa(idcorporate)
	strweek := strconv.Itoa(week)

	querystart := `
	SELECT idpedagangkios, namapedagang, noreg, idtagihankategori,
		uraiantagihankategori, sum(poin), idcorporate,
		cid, namacorporate, date_part('week', tanggaltransaksi::date) AS week,
		EXTRACT(MONTH FROM min(tanggaltransaksi)) AS thisweekmonth,
		EXTRACT(MONTH FROM min(tanggaltransaksi) + interval '7' day) AS nextweekmonth,
		EXTRACT(YEAR FROM min(tanggaltransaksi)) AS year
	FROM pedagangkiospoin
	WHERE deleted_at IS NULL AND idtagihankategori = 1`

	queryend := `
	GROUP BY week, idpedagangkios, namapedagang, noreg, idtagihankategori,
		uraiantagihankategori, idcorporate, cid, namacorporate
	ORDER BY idpedagangkios, week ASC, idtagihankategori DESC`

	if idcorporate != 0 {
		querystart += " AND idcorporate = " + stridcorporate
	}
	if week != 0 {
		querystart += " AND date_part('week', tanggaltransaksi::date) = " + strweek
	}

	querystart += queryend
	rows, err := ctx.RepoDB.DB.Query(querystart)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var val models.PedagangKiosGradingWeekly
		err := rows.Scan(&val.IDPedagangKios, &val.NamaPedagang, &val.NoReg, &val.IDTagihanKategori,
			&val.UraianTagihanKategori, &val.Poin, &val.IDCorporate, &val.CID, &val.NamaCorporate,
			&val.Week, &val.BulanMingguIni, &val.BulanMingguDepan, &val.Year)
		if err != nil {
			return result, err
		}
		result = append(result, val)
	}

	return result, nil
}

// AddPedagangKiosGradingWeekly
func (ctx pedagangKiosGradingRepository) AddPedagangKiosGradingWeekly(req []models.ResponseFindPedagangKiosGradingWeekly) (bool, error) {
	var result bool

	sqlStr := `INSERT INTO pedagangkiosgradingmingguan (created_at, namacorporate, noreg, year, week, 
		poin, grade, namapedagang, kodegrading) VALUES `
	vals := []interface{}{}
	t := time.Now()
	dbTime := t.Format("2006-01-02 15:04:05")
	for _, row := range req {
		sqlStr += "(?, ?, ?, ?, ?, ?, ?, ?, ?),"

		vals = append(vals, dbTime, row.NamaCorporate, row.NoReg, row.Year, row.Week,
			row.Poin, row.Grade, row.NamaPedagang, row.KodeGrading)
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

func (ctx pedagangKiosGradingRepository) FindWeekPoinBonus() ([]models.WeekPoinBonus, error) {
	var result []models.WeekPoinBonus

	rows, err := ctx.RepoDB.DB.Query(`
	SELECT idcorporate, date_part('week', tanggal::date) AS week, count(idcorporate)
	FROM tanggallibur
	WHERE deleted_at IS NULL
	GROUP BY idcorporate, week
	ORDER BY idcorporate ASC`)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var prd models.WeekPoinBonus
		rows.Scan(&prd.IDCorporate, &prd.Week, &prd.BonusPoin)
		result = append(result, prd)
	}
	err = rows.Err()
	if err != nil {
		return result, err
	}
	return result, nil
}
