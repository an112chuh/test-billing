package routes

import (
	"work/handlers"

	"github.com/gorilla/mux"
)

func GetAllHandlers(r *mux.Router) {
	r.HandleFunc("/api/new", handlers.NewBillHandler)
	r.HandleFunc("/api/confirm", handlers.ChangeStatusHandler)
	r.HandleFunc("/api/bill", handlers.GetStatusByIDHandler)
	r.HandleFunc("/api/user_profile", handlers.GetBillsByUserHandler)
	r.HandleFunc("/api/cancel", handlers.CancelBillHandler)
}
