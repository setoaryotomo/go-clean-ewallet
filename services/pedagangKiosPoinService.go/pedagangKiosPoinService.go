package pedagangKiosPoinService

import (
	"math"
	"net/http"
	"sample/constans"
	"sample/helpers"
	"sample/models"
	"sample/services"
	"strconv"
	"time"

	"github.com/labstack/echo"
)

type pedagangKiosPoinService struct {
	Service services.UsecaseService
}

// NewPedagangKiosPoinService
func NewPedagangKiosPoinService(service services.UsecaseService) pedagangKiosPoinService {
	return pedagangKiosPoinService{
		Service: service,
	}
}

// PedagangKiosPoinDaily ..
func (svc pedagangKiosPoinService) AddPedagangKiosPoinDaily(ctx echo.Context) error {
	var result models.Response
	var request []models.PedagangKiosPoin

	timeFrom := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC)
	timeTo := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 23, 59, 59, 0, time.UTC)
	// timeFrom := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	// timeTo := time.Date(2021, 2, 20, 23, 59, 59, 0, time.UTC)
	days := timeTo.Sub(timeFrom).Hours() / 24
	temp := int(math.Ceil(days))

	for i := 0; i < temp; i++ {
		a := timeFrom.AddDate(0, 0, i).Format("2006-01-02")

		bulkdata, err := svc.Service.PedagangKiosPoinRepo.FindPedagangKiosDataByDate(a)
		if err != nil {
			result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
			return ctx.JSON(http.StatusBadRequest, result)
		}

		for _, raw := range bulkdata {
			var poin int
			switch {
			case raw.IDTagihanKategori == 1:
				poin = 1
			case raw.IDTagihanKategori == 2:
				poin = 30
			case raw.IDTagihanKategori == 3:
				poin = 365
			}
			pedagangkiospoin := models.PedagangKiosPoin{
				IDPedagangKios:         raw.IDPedagangKios,
				NamaPedagang:           raw.NamaPedagang,
				NoReg:                  raw.NoReg,
				IDTagihanKategori:      raw.IDTagihanKategori,
				UraianTagihanKategori:  raw.UraianTagihanKategori,
				TanggalTransaksi:       a,
				PedagangKiosPoinUnique: a + "/" + strconv.Itoa(raw.IDPedagangKios) + "/" + strconv.Itoa(raw.IDTagihanKategori),
				IDCorporate:            raw.IDCorporate,
				CID:                    raw.CID,
				NamaCorporate:          raw.NamaCorporate,
				Poin:                   poin,
			}

			request = append(request, pedagangkiospoin)
		}

		if len(request) != 0 {
			_, err := svc.Service.PedagangKiosPoinRepo.AddPedagangKiosPoin(request)
			if err != nil {
				result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
				return ctx.JSON(http.StatusBadRequest, result)
			}
		}
		request = nil

	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, constans.EMPTY_VALUE, true)

	return ctx.JSON(http.StatusOK, result)
}

// AddPedagangKiosPoinInisiasi ..
func (svc pedagangKiosPoinService) AddPedagangKiosPoinInisiasi(ctx echo.Context) error {
	helpers.LOG("PEDAGANGKIOSPOIN", "OK", true)
	var result models.Response
	var request []models.PedagangKiosPoin

	// timeFrom := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC)
	timeTo := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 23, 59, 59, 0, time.UTC)
	timeFrom := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	// timeTo := time.Date(2021, 2, 20, 23, 59, 59, 0, time.UTC)
	days := timeTo.Sub(timeFrom).Hours() / 24
	temp := int(math.Ceil(days))

	for i := 0; i < temp; i++ {
		a := timeFrom.AddDate(0, 0, i).Format("2006-01-02")

		bulkdata, err := svc.Service.PedagangKiosPoinRepo.FindPedagangKiosDataByDate(a)
		if err != nil {
			result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
			return ctx.JSON(http.StatusBadRequest, result)
		}

		for _, raw := range bulkdata {
			var poin int
			switch {
			case raw.IDTagihanKategori == 1:
				poin = 1
			case raw.IDTagihanKategori == 2:
				poin = 30
			case raw.IDTagihanKategori == 3:
				poin = 365
			}
			pedagangkiospoin := models.PedagangKiosPoin{
				IDPedagangKios:         raw.IDPedagangKios,
				NamaPedagang:           raw.NamaPedagang,
				NoReg:                  raw.NoReg,
				IDTagihanKategori:      raw.IDTagihanKategori,
				UraianTagihanKategori:  raw.UraianTagihanKategori,
				TanggalTransaksi:       a,
				PedagangKiosPoinUnique: a + "/" + strconv.Itoa(raw.IDPedagangKios) + "/" + strconv.Itoa(raw.IDTagihanKategori),
				IDCorporate:            raw.IDCorporate,
				CID:                    raw.CID,
				NamaCorporate:          raw.NamaCorporate,
				Poin:                   poin,
			}

			request = append(request, pedagangkiospoin)
		}

		// if len(request) != 0 {
		// 	_, err := svc.Service.PedagangKiosPoinRepo.AddPedagangKiosPoin(request)
		// 	if err != nil {
		// 		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		// 		return ctx.JSON(http.StatusBadRequest, result)
		// 	}
		// }

		var batch []models.PedagangKiosPoin
		for l := 0; l < len(request); l++ {
			var isibatch models.PedagangKiosPoin
			isibatch = request[l]
			// fmt.Println(isibatch)
			batch = append(batch, isibatch)

			if l > 0 {
				if l%100 == 0 || l == len(request) {
					_, err = svc.Service.PedagangKiosPoinRepo.AddPedagangKiosPoin(batch)
					if err != nil {
						result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, err.Error(), nil)
						return ctx.JSON(http.StatusBadRequest, result)
					}
					batch = nil
				}
			}

		}

		request = nil

	}

	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, constans.EMPTY_VALUE, true)

	return ctx.JSON(http.StatusOK, result)
}
