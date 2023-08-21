package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"
	"work/check"
	"work/config"
	"work/result"
)

type NewBillInput struct {
	ID       int     `json:"id"`
	EMail    string  `json:"email"`
	Sum      float64 `json:"sum"`
	Currency string  `json:"currency"`
}

type UpdateBillStatus struct {
	ID     int    `json:"id"`
	Status int    `json:"status"`
	ApiKey string `json:"api_key"`
}

type BillStatuses struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
}

type BillList struct {
	IDUser *int           `json:"id_user,omitempty"`
	Email  *string        `json:"mail,omitempty"`
	Bills  []BillStatuses `json:"bill_statuses"`
}

type BillCancel struct {
	ID int `json:"id"`
}

var Statuses [5]string = [5]string{"НОВЫЙ", "УСПЕХ", "НЕУСПЕХ", "ОШИБКА", "ОТМЕНЕН"}

/*
var NEW = Statuses[1]
var NEW = Statuses[1]
var NEW = Statuses[1]
var NEW = Statuses[1]
var NEW = Statuses[1] */

func NewBillHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	var data NewBillInput
	b, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(b))
	err := json.Unmarshal(b, &data)
	if err != nil {
		result.ErrorServer(r, err)
		res = result.SetErrorResult(result.UnknownError)
		result.ReturnJSON(w, &res)
		return
	}
	res = billHandler(r, data)
	result.ReturnJSON(w, &res)
}

func ChangeStatusHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	var data UpdateBillStatus
	b, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(b))
	err := json.Unmarshal(b, &data)
	if err != nil {
		result.ErrorServer(r, err)
		res = result.SetErrorResult(result.UnknownError)
		result.ReturnJSON(w, &res)
		return
	}
	res = changeStatusHandler(r, data)
	result.ReturnJSON(w, &res)
}

func GetStatusByIDHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	vars := r.URL.Query()
	if len(vars["id"]) == 0 {
		res = result.SetErrorResult(`Неверный параметр id платежа`)
		result.ReturnJSON(w, &res)
		return
	}
	id := vars["id"][0]
	ID, err := strconv.Atoi(id)
	if err != nil {
		result.ErrorServer(r, err)
		res = result.SetErrorResult(`Неверный параметр id платежа`)
		result.ReturnJSON(w, &res)
		return
	}
	res = getStatus(r, ID)
	result.ReturnJSON(w, &res)
}

func GetBillsByUserHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	vars := r.URL.Query()
	if len(vars["id"]) > 0 {
		id := vars["id"][0]
		ID, err := strconv.Atoi(id)
		if err != nil {
			res = result.SetErrorResult(`Неверный параметр id пользователя`)
			result.ReturnJSON(w, &res)
			return
		}
		res = getBill(r, &ID, nil)
	} else if len(vars["mail"]) > 0 {
		email := vars["mail"][0]
		res = getBill(r, nil, &email)
	} else {
		res = result.SetErrorResult(`Неверные параметры запроса(требуется либо id, либо mail)`)
	}
	result.ReturnJSON(w, &res)
}

func CancelBillHandler(w http.ResponseWriter, r *http.Request) {
	var res result.ResultInfo
	var data BillCancel
	b, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewReader(b))
	err := json.Unmarshal(b, &data)
	if err != nil {
		result.ErrorServer(r, err)
		res = result.SetErrorResult(result.UnknownError)
		result.ReturnJSON(w, &res)
		return
	}
	res = billCancel(r, data)
	result.ReturnJSON(w, &res)
}

func billHandler(r *http.Request, data NewBillInput) (res result.ResultInfo) {
	isValid, err := check.IsUserValid(data.ID, &data.EMail)
	if err != nil {
		result.ErrorServer(r, err)
		res = result.SetErrorResult(result.UnknownError)
		return
	}
	if !isValid {
		res = result.SetErrorResult("Пользователь не найден")
		return
	}
	isValid = check.IsCurrencyValid(data.Currency)
	if !isValid {
		res = result.SetErrorResult("Код валюты должен содержать 3 заглавных буквы")
		return
	}
	db := config.GetConnection()
	query := `INSERT INTO transactions (id_user, sum, currency, created_at, updated_at, status) VALUES ($1, $2, $3, $4, $5, $6)`
	t := time.Now()
	params := []any{data.ID, data.Sum, data.Currency, t, t, createRandomStatus()}
	_, err = db.Exec(query, params...)
	if err != nil {
		result.ErrorServer(r, err)
		res = result.SetErrorResult(result.UnknownError)
		return
	}
	res.Done = true
	return
}

func changeStatusHandler(r *http.Request, data UpdateBillStatus) (res result.ResultInfo) {
	isValid := check.IsNewBillStatusValid(data.Status)
	if !isValid {
		res = result.SetErrorResult("Неверный новый статус")
		return
	}
	isValid = check.IsBillValid(data.ID)
	if !isValid {
		res = result.SetErrorResult("Такой сделки не существует")
		return
	}
	isValid, err := check.IsApiKeyValid(data.ApiKey)
	if err != nil {
		result.ErrorServer(r, err)
		res = result.SetErrorResult(result.UnknownError)
		return
	}
	if !isValid {
		res = result.SetErrorResult("Неверный ключ")
		return
	}
	status, err := check.CurrentStatus(data.ID)
	if err != nil {
		result.ErrorServer(r, err)
		res = result.SetErrorResult(result.UnknownError)
		return
	}
	if status == 1 {
		db := config.GetConnection()
		query := `UPDATE transactions SET status = $1 WHERE id = $2`
		params := []any{data.Status, data.ID}
		_, err := db.Exec(query, params...)
		if err != nil {
			result.ErrorServer(r, err)
			res = result.SetErrorResult(result.UnknownError)
			return
		}
		res.Done = true
		return
	}
	res = result.SetErrorResult("Невозможно изменить состояние")
	return
}

func getStatus(r *http.Request, ID int) (res result.ResultInfo) {
	status, err := check.CurrentStatus(ID)
	if err != nil {
		result.ErrorServer(r, err)
		res = result.SetErrorResult(result.UnknownError)
		return
	}
	res.Done = true
	res.Items = BillStatuses{ID: ID, Status: Statuses[status-1]}
	return
}

func getBillByID(r *http.Request, ID int) (query string, params []any) {
	query = `SELECT id, status FROM transactions WHERE id_user = $1`
	params = []any{ID}
	return
}

func getBillByEmail(r *http.Request, Email string) (query string, params []any) {
	query = `SELECT transactions.id, status FROM transactions
		INNER JOIN users on users.id = transactions.id_user
		WHERE email = $1`
	params = []any{Email}
	return
}

func getBill(r *http.Request, ID *int, Email *string) (res result.ResultInfo) {
	var data BillList
	db := config.GetConnection()
	var query string
	var params []any
	if ID == nil {
		query, params = getBillByEmail(r, *Email)
	} else if Email == nil {
		query, params = getBillByID(r, *ID)
	} else {
		result.ErrorServer(r, errors.New("параметры ID и Email пусты"))
		res = result.SetErrorResult(result.UnknownError)
		return
	}
	rows, err := db.Query(query, params...)
	if err != nil {
		result.ErrorServer(r, err)
		res = result.SetErrorResult(result.UnknownError)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var b BillStatuses
		var status int
		err := rows.Scan(&b.ID, &status)
		if err != nil {
			result.ErrorServer(r, err)
			res = result.SetErrorResult(result.UnknownError)
			return
		}
		b.Status = Statuses[status-1]
		data.Bills = append(data.Bills, b)
	}
	data.Email = Email
	data.IDUser = ID
	res.Done = true
	res.Items = data
	return
}

func billCancel(r *http.Request, data BillCancel) (res result.ResultInfo) {
	status, err := check.CurrentStatus(data.ID)
	if err != nil {
		result.ErrorServer(r, err)
		res = result.SetErrorResult(result.UnknownError)
		return
	}
	if status == 1 {
		db := config.GetConnection()
		query := `UPDATE transactions SET status = 5 WHERE id = $1`
		params := []any{data.ID}
		_, err := db.Exec(query, params...)
		if err != nil {
			result.ErrorServer(r, err)
			res = result.SetErrorResult(result.UnknownError)
			return
		}
		res.Done = true
		return
	}
	res = result.SetErrorResult("Невозможно отменить платёж")
	return
}

func createRandomStatus() int {
	value := rand.Intn(100)
	if value <= 9 {
		return 4
	}
	return 1
}
