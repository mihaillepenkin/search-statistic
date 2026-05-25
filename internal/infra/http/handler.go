package handler

import (
	"encoding/json"
	"net/http"
	"search_statistics/internal/usecase"
	"strconv"

	"github.com/gorilla/mux"
)

type Handler struct {
	srv usecase.Service
}

func MakeHandler(serv usecase.Service) *Handler {
	handler := new(Handler)
	handler.srv = serv
	return handler
}

func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (h *Handler) GetTopHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	nStr := vars["n"]

	if nStr == "" {
		w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "Missing number_of_positions path parameter"})
		return
	}

	n, err := strconv.Atoi(nStr)
	if (err != nil || n <= 0) {
		w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "invalid number_of_positions path parameter"})
		return
	}

	var input usecase.GetTopDTO
	input.NumberOfPositions = n
	output := h.srv.GetTop(&input)
	if output.Msg != "" {
		w.WriteHeader(500)
        json.NewEncoder(w).Encode(map[string]string{"error": output.Msg})
        return
	}
	json.NewEncoder(w).Encode(output)
}

func (h *Handler) AddInStopListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var input usecase.AddInStopListDTO
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	output := h.srv.AddInStopList(&input)
	if output.Msg != "" {
		w.WriteHeader(500)
        json.NewEncoder(w).Encode(map[string]string{"error": output.Msg})
        return
	}
	json.NewEncoder(w).Encode(output)
}

func (h *Handler) DeleteFromStopList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var input usecase.DeleteFromStopListDTO
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	output := h.srv.DeleteFromStopList(&input)
	if output.Msg != "" {
		w.WriteHeader(500)
        json.NewEncoder(w).Encode(map[string]string{"error": output.Msg})
        return
	}
	json.NewEncoder(w).Encode(output)
}

