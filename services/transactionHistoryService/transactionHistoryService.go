package transactionHistoryService

import (
	"fmt"
	"net/http"
	"sample/constans"
	"sample/helpers"
	"sample/models"
	"sample/services"
	"sample/utils"

	"github.com/labstack/echo"
)

type transactionHistoryService struct {
	Service services.UsecaseService
}

func NewTransactionHistoryService(service services.UsecaseService) transactionHistoryService {
	return transactionHistoryService{
		Service: service,
	}
}

// TransactionHistoryListV2 mendapatkan list history transaksi dengan filter dan pagination
func (svc transactionHistoryService) TransactionHistoryListV2(ctx echo.Context) error {
	var (
		request                 models.RequestTransactionHistoryList
		responseTransactionList []models.ResponseTransactionHistoryListV2
		result                  models.Response
		resultValue             models.ResponseTransactionHistoryV2
		resCountAndSummaries    models.ResultDataTableTransactionCountAndSummaries
		serviceName             = "TransactionService.TransactionHistoryListV2"
		err                     error
	)

	// Generate reference number untuk tracking request ini
	referenceNo := utils.GenerateShortReferenceNo()

	// Validate request struct
	if err := helpers.BindValidateStruct(ctx, &request); err != nil {
		utils.LogError(serviceName, referenceNo, "BindValidateStruct", err)
		result = helpers.ResponseJSON(false, constans.VALIDATE_ERROR_CODE, err.Error(), nil)
		return ctx.JSON(http.StatusBadRequest, result)
	}

	accNo := request.AccountNumber
	if accNo == "" {
		accNo = "ALL_ACCOUNTS"
	}

	// Log dengan reference number
	utils.LogInfo(serviceName, referenceNo, "Request received",
		fmt.Sprintf("RefNo: %s, AccountNo: %s, StartDate: %s, EndDate: %s, PageNumber: %d, PageSize: %d",
			referenceNo, accNo, request.StartDate, request.EndDate, request.PageNumber, request.PageSize))

	// Validate Account if provided
	if request.AccountNumber != "" {
		_, err := svc.Service.AccountRepo.FindAccountByNumber(request.AccountNumber)
		if err != nil {
			utils.LogError(serviceName, referenceNo, "FindAccountByNumber", err)
			result = helpers.ResponseJSON(false, constans.DATA_NOT_FOUND_CODE, "Account not found", nil)
			return ctx.JSON(http.StatusNotFound, result)
		}
	}

	// Get count and summaries
	if request.Draw == 1 {
		resCountAndSummaries, err = svc.Service.TransactionRepo.DataCountAndSumTransactionListByIndex(false, request)
		if err != nil {
			utils.LogError(serviceName, referenceNo, "DataCountAndSumTransactionListByIndex (countOnly false)", err)
		}
	} else {
		resCountAndSummaries, err = svc.Service.TransactionRepo.DataCountAndSumTransactionListByIndex(true, request)
		if err != nil {
			utils.LogError(serviceName, referenceNo, "DataCountAndSumTransactionListByIndex (countOnly true)", err)
		}
	}

	// Get transaction list
	resListTransaction, err := svc.Service.TransactionRepo.DataGetTransactionListByIndex(request)
	if err != nil {
		utils.LogError(serviceName, referenceNo, "DataGetTransactionListByIndex", err)
		result = helpers.ResponseJSON(false, constans.SYSTEM_ERROR_CODE, "Failed to get transaction list: "+err.Error(), nil)
		return ctx.JSON(http.StatusInternalServerError, result)
	}

	// Build response list
	for _, tx := range resListTransaction {
		responseTransactionList = append(responseTransactionList, models.ResponseTransactionHistoryListV2{
			ID:            tx.ID,
			AccountID:     tx.AccountID,
			AccountNumber: tx.AccountNumber,
			AccountName:   tx.AccountName,
			// SourceNumber:      tx.SourceNumber,
			// BeneficiaryNumber: tx.BeneficiaryNumber,
			TransactionType: tx.TransactionType,
			Amount:          tx.Amount,
			TransactionTime: tx.TransactionTime.Format(constans.LAYOUT_TIMESTAMP),
			CreatedAt:       tx.CreatedAt.Format(constans.LAYOUT_TIMESTAMP),
		})
	}

	// Build final response
	resultValue = models.ResponseTransactionHistoryV2{
		ReferenceNo:     referenceNo,
		RecordsFiltered: len(resListTransaction),
		RecordsTotal:    int(resCountAndSummaries.Count),
		Value:           responseTransactionList,
	}

	utils.LogInfo(serviceName, referenceNo, "TransactionHistoryListV2.Success",
		fmt.Sprintf("RefNo: %s, Retrieved %d transactions (Total: %d)", referenceNo, len(responseTransactionList), resCountAndSummaries.Count))

	// Response TransactionHistory
	result = helpers.ResponseJSON(true, constans.SUCCESS_CODE, "Transaction history retrieved successfully", resultValue)
	return ctx.JSON(http.StatusOK, result)
}
