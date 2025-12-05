package pedagangKiosGradingService

import (
	"net/http"
	"sample/constans"
	"sample/helpers"
	"sample/models"
	"sample/services"
	"strconv"
	"time"

	"github.com/labstack/echo"
)

type pedagangKiosGradingService struct {
	Service services.UsecaseService
}

// NewPedagangKiosGradingService
func NewPedagangKiosGradingService(service services.UsecaseService) pedagangKiosGradingService {
	return pedagangKiosGradingService{
		Service: service,
	}
}

func (svc pedagangKiosGradingService) AddPedagangKiosGradingWeekly(ctx echo.Context) error {
	var result models.Response
	var response []models.ResponseFindPedagangKiosGradingWeekly

	tempidpedagangkios := 0
	tempweek := 0
	temppoinbulanan := 0
	_, thiswek := time.Now().ISOWeek()
	IDCorporate, _ := strconv.Atoi(ctx.Param("idcorporate"))

	resPedagangKiosGrading, err := svc.Service.PedagangKiosGradingRepo.FindPedagangKiosGradingWeekly(IDCorporate, 0)
	if err != nil {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	resweekPoinBonus, err := svc.Service.PedagangKiosGradingRepo.FindWeekPoinBonus()
	if err != nil {
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	for i, pdkpoin := range resPedagangKiosGrading {
		if tempidpedagangkios == 0 { // Mengisi variable temp jika kosong
			tempidpedagangkios = pdkpoin.IDPedagangKios
		}
		if tempweek == 0 { // Mengisi variable temp jika kosong
			tempweek = pdkpoin.Week
		}
		if pdkpoin.IDTagihanKategori == 2 && temppoinbulanan == 0 { // Mengisi variable temp bulanan jika kosong
			temppoinbulanan = pdkpoin.Poin
		}
		if pdkpoin.IDPedagangKios == tempidpedagangkios {
			if pdkpoin.IDTagihanKategori == 1 {
				if tempweek != pdkpoin.Week {
					for tempweek+1 != pdkpoin.Week {
						tempweek += 1
						resPedagangKiosGrading[i].Week = tempweek
						resPedagangKiosGrading[i].Poin = 0
						a, newpoinbulanan := repeatedAppendGrading(resPedagangKiosGrading[i], resweekPoinBonus, temppoinbulanan)
						temppoinbulanan = newpoinbulanan
						response = append(response, a)
					}
				}
				a, newpoinbulanan := repeatedAppendGrading(pdkpoin, resweekPoinBonus, temppoinbulanan)
				temppoinbulanan = newpoinbulanan
				response = append(response, a)
				tempweek = pdkpoin.Week
			}
		} else {
			// Digunakan untuk case jika orang tersebut tidak membayar hingga minggu ini
			for tempweek != thiswek {
				tempweek += 1
				resPedagangKiosGrading[i-1].Week = tempweek
				resPedagangKiosGrading[i-1].Poin = 0
				a, newpoinbulanan := repeatedAppendGrading(resPedagangKiosGrading[i-1], resweekPoinBonus, temppoinbulanan)
				temppoinbulanan = newpoinbulanan
				response = append(response, a)
			}
			a, newpoinbulanan := repeatedAppendGrading(pdkpoin, resweekPoinBonus, temppoinbulanan)
			temppoinbulanan = newpoinbulanan
			response = append(response, a)
			tempidpedagangkios = pdkpoin.IDPedagangKios
			tempweek = pdkpoin.Week
			temppoinbulanan = 0
		}
	}

	var batchresponse []models.ResponseFindPedagangKiosGradingWeekly

	for l := 0; l < len(response); l++ {
		var isibatch models.ResponseFindPedagangKiosGradingWeekly
		isibatch = response[l]
		batchresponse = append(batchresponse, isibatch)

		if l > 0 {
			if l%100 == 0 || l == len(response) {
				_, err = svc.Service.PedagangKiosGradingRepo.AddPedagangKiosGradingWeekly(batchresponse)
				if err != nil {
					result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, err.Error(), nil)
					return ctx.JSON(http.StatusBadRequest, result)
				}
				batchresponse = nil
			}
		}

	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, constans.EMPTY_VALUE, response)
	return ctx.JSON(http.StatusOK, result)
}

func repeatedAppendGrading(data models.PedagangKiosGradingWeekly, weekpoinbonus []models.WeekPoinBonus, bonuspoinbulanan int) (models.ResponseFindPedagangKiosGradingWeekly, int) {
	var result models.ResponseFindPedagangKiosGradingWeekly
	result.NamaPedagang = data.NamaPedagang
	result.NoReg = data.NoReg
	result.NamaCorporate = data.NamaCorporate
	result.Week = data.Week
	result.Year = data.Year
	result.KodeGrading = "WG" + "-" + data.NoReg + "/" + strconv.Itoa(data.Year) + "/" + strconv.Itoa(data.Week)

	// start, end := WeekRange(data.Year, data.Week)
	// if start.Month() != end.Month() || end.AddDate(0, 0, 1).Month() != end.Month() {
	// 	// Check apakah betul minggu ini adalah minggu terakhir
	// 	// Jika betul maka menambahkan bonuspoinbulanan, setelah digunakan kosongkan bonuspoinbulanan
	// 	result.Poin = data.Poin + FindWeekPoinBonus(weekpoinbonus, data.IDCorporate, data.Week) + bonuspoinbulanan
	// 	switch {
	// 	case result.Poin == 37: // Bayar bulanan + harian komplit
	// 		result.Grade = "A"
	// 	case result.Poin >= 33: // Bayar bulanan + harian kurang
	// 		result.Grade = "B"
	// 	case result.Poin >= 30: // Bayar bulanan + harian kosong
	// 		result.Grade = "C"
	// 	case result.Poin >= 7 && result.Week == 8 && result.Year == 2021: // Bulan Februari belum ada tagihan bulanan
	// 		result.Grade = "A"
	// 	case result.Poin >= 3 && result.Week == 8 && result.Year == 2021: // Bulan Februari belum ada tagihan bulanan
	// 		result.Grade = "B"
	// 	case result.Poin == 7: // Tidak bayar bulanan + harian komplit
	// 		result.Grade = "B"
	// 	default:
	// 		result.Grade = "C" // Tidak bayar bulanan + harian kurang
	// 	}
	// 	bonuspoinbulanan = 0
	// } else {
	// Bukan minggu terakhir
	result.Poin = data.Poin + checkWeekPoinBonus(weekpoinbonus, data.IDCorporate, data.Week)
	switch {
	case result.Poin == 7: // Bayar harian komplit
		result.Grade = "A"
	case result.Poin >= 4: // Bayar harian kurang sedikit
		result.Grade = "B"
	default:
		result.Grade = "C" // Bayar harian kurang banyak
	}
	// }

	return result, bonuspoinbulanan
}

//CheckWeekPoinBonus mencari apakah pasar dan minggu tersebut terdapat hari libur
func checkWeekPoinBonus(slice []models.WeekPoinBonus, idcorporate int, week int) int {
	for _, data := range slice {
		if data.IDCorporate == idcorporate && data.Week == week {
			return data.BonusPoin
		}
	}
	return 0
}
