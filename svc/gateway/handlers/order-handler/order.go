package orderhandler

import (
	"context"
	"errors"
	orderstructs "gateway/handlers/order-handler/order-structs"
	"gateway/types"
	"gateway/util"
	"net/http"
	"pkg/json"
	orderv1 "proto/out/order/v1"
	"strconv"
	"time"
)

type Handler struct {
	client orderv1.OrderServiceClient
}

func NewHandler(client orderv1.OrderServiceClient) *Handler {
	return &Handler{
		client: client,
	}
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	token, ok := r.Context().Value(types.TokenKey).(string)
	if !ok {
		http.Error(w, "token not found", http.StatusUnauthorized)
		return
	}
	orderID := r.PathValue("id")
	if !util.IsValidUUID(orderID) {
		http.Error(w, "invalid order id", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()
	res, err := h.client.FetchOrder(ctx, &orderv1.FetchOrderRequest{
		Token:   token,
		OrderId: orderID,
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, "request timed out", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := orderstructs.OrderResponse{
		OrderID:   orderID,
		Total:     res.TotalPrice,
		Items:     make([]orderstructs.Item, 0, len(res.Items)),
		CreatedAt: res.CreatedAt,
	}
	for _, item := range res.Items {
		resp.Items = append(resp.Items, orderstructs.Item{
			ItemID:   item.ItemId,
			Name:     item.Name,
			Quantity: int(item.Quantity),
			Price:    item.Price,
		})
	}
	json.Write(w, http.StatusOK, &resp)
}

func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	token, ok := r.Context().Value(types.TokenKey).(string)
	if !ok {
		http.Error(w, "token not found", http.StatusUnauthorized)
		return
	}
	pageStr := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		http.Error(w, "page is not a number", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()
	res, err := h.client.FetchOrders(ctx, &orderv1.FetchOrdersRequest{
		Token: token,
		Page:  int32(page),
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			http.Error(w, "request timed out", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := orderstructs.OrdersResponse{
		Orders: make([]orderstructs.Order, 0, len(res.Orders)),
	}
	for _, order := range res.Orders {
		resp.Orders = append(resp.Orders, orderstructs.Order{
			OrderID:   order.OrderId,
			Total:     order.TotalPrice,
			Status:    order.Status,
			CreatedAt: order.CreatedAt,
		})
	}
	json.Write(w, http.StatusOK, &resp)
}

func (h *Handler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	token, ok := r.Context().Value(types.TokenKey).(string)
	if !ok {
		http.Error(w, "token not found", http.StatusUnauthorized)
		return
	}
	orderID := r.PathValue("id")
	if !util.IsValidUUID(orderID) {
		http.Error(w, "invalid order id", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()
	_, err := h.client.CancelOrder(ctx, &orderv1.CancelOrderRequest{
		Token:   token,
		OrderId: orderID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
