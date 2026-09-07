package handlers

import (
	orderhandler "gateway/handlers/order-handler"
	"gateway/middleware"
	"net/http"
	orderv1 "proto/out/order/v1"
)

func RegisterOrderHandler(mux *http.ServeMux, cl orderv1.OrderServiceClient, mw *middleware.Middleware) {
	h := orderhandler.NewHandler(cl)
	mux.HandleFunc("GET /orders/{id}", mw.SetUserID(http.HandlerFunc(h.GetOrder)))
	mux.HandleFunc("GET /orders", mw.SetUserID(http.HandlerFunc(h.GetOrders)))
	mux.HandleFunc("PATCH /orders/{id}", mw.SetUserID(http.HandlerFunc(h.CancelOrder)))
}
