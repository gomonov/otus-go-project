package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gomonov/otus-go-project/internal/domain"
)

type CreateSubnetRequest struct {
	CIDR string `json:"cidr"`
}

type DeleteSubnetRequest struct {
	CIDR string `json:"cidr"`
}

type SubnetResponse struct {
	ListType domain.ListType `json:"listType"`
	CIDR     string          `json:"cidr"`
}

type SubnetsListResponse struct {
	Subnets []SubnetResponse `json:"subnets"`
	Count   int              `json:"count"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (s *Server) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", s.rootHandler)
	mux.HandleFunc("/blacklist", s.blacklistHandler)
	mux.HandleFunc("/whitelist", s.whitelistHandler)
	mux.HandleFunc("/auth", s.authHandler)

	return mux
}

func (s *Server) rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	apiInfo := map[string]interface{}{
		"service": "Network Lists API",
		"version": "1.0",
		"endpoints": []map[string]string{
			{
				"method":      "GET",
				"path":        "/blacklist",
				"description": "Get all subnets from blacklist",
			},
			{
				"method":      "POST",
				"path":        "/blacklist",
				"description": "Add subnet to blacklist",
			},
			{
				"method":      "DELETE",
				"path":        "/blacklist",
				"description": "Remove subnet from blacklist",
			},
			{
				"method":      "GET",
				"path":        "/whitelist",
				"description": "Get all subnets from whitelist",
			},
			{
				"method":      "POST",
				"path":        "/whitelist",
				"description": "Add subnet to whitelist",
			},
			{
				"method":      "DELETE",
				"path":        "/whitelist",
				"description": "Remove subnet from whitelist",
			},
		},
	}

	json.NewEncoder(w).Encode(apiInfo)
}

func (s *Server) blacklistHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getBlacklistHandler(w)
	case http.MethodPost:
		s.addToBlacklistHandler(w, r)
	case http.MethodDelete:
		s.removeFromBlacklistHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) whitelistHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getWhitelistHandler(w)
	case http.MethodPost:
		s.addToWhitelistHandler(w, r)
	case http.MethodDelete:
		s.removeFromWhitelistHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) getBlacklistHandler(w http.ResponseWriter) {
	subnets, err := s.app.GetSubnetsByListType(domain.Blacklist)
	if err != nil {
		s.sendError(w, fmt.Sprintf("Failed to get blacklist: %v", err), http.StatusInternalServerError)
		return
	}

	response := s.convertSubnetsToResponse(subnets)
	s.sendJSON(w, response, http.StatusOK)
}

func (s *Server) getWhitelistHandler(w http.ResponseWriter) {
	subnets, err := s.app.GetSubnetsByListType(domain.Whitelist)
	if err != nil {
		s.sendError(w, fmt.Sprintf("Failed to get whitelist: %v", err), http.StatusInternalServerError)
		return
	}

	response := s.convertSubnetsToResponse(subnets)
	s.sendJSON(w, response, http.StatusOK)
}

func (s *Server) addToBlacklistHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateSubnetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.CIDR == "" {
		s.sendError(w, "CIDR is required", http.StatusBadRequest)
		return
	}

	subnet := &domain.Subnet{
		ListType: domain.Blacklist,
		CIDR:     req.CIDR,
	}

	if err := s.app.CreateSubnet(subnet); err != nil {
		s.sendError(w, fmt.Sprintf("Failed to add to blacklist: %v", err), http.StatusInternalServerError)
		return
	}

	response := SubnetResponse{
		ListType: subnet.ListType,
		CIDR:     subnet.CIDR,
	}

	s.sendJSON(w, response, http.StatusCreated)
}

func (s *Server) addToWhitelistHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateSubnetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.CIDR == "" {
		s.sendError(w, "CIDR is required", http.StatusBadRequest)
		return
	}

	subnet := &domain.Subnet{
		ListType: domain.Whitelist,
		CIDR:     req.CIDR,
	}

	if err := s.app.CreateSubnet(subnet); err != nil {
		s.sendError(w, fmt.Sprintf("Failed to add to whitelist: %v", err), http.StatusInternalServerError)
		return
	}

	response := SubnetResponse{
		ListType: subnet.ListType,
		CIDR:     subnet.CIDR,
	}

	s.sendJSON(w, response, http.StatusCreated)
}

func (s *Server) removeFromBlacklistHandler(w http.ResponseWriter, r *http.Request) {
	var req DeleteSubnetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.CIDR == "" {
		s.sendError(w, "CIDR is required", http.StatusBadRequest)
		return
	}

	if err := s.app.DeleteSubnet(domain.Blacklist, req.CIDR); err != nil {
		if errors.Is(err, domain.ErrSubnetNotFound) {
			s.sendError(w, "Subnet not found in blacklist", http.StatusNotFound)
		} else {
			s.sendError(w, fmt.Sprintf("Failed to remove from blacklist: %v", err), http.StatusInternalServerError)
		}
		return
	}

	s.sendJSON(w, map[string]string{"message": "Subnet removed from blacklist successfully"}, http.StatusOK)
}

func (s *Server) removeFromWhitelistHandler(w http.ResponseWriter, r *http.Request) {
	var req DeleteSubnetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.CIDR == "" {
		s.sendError(w, "CIDR is required", http.StatusBadRequest)
		return
	}

	if err := s.app.DeleteSubnet(domain.Whitelist, req.CIDR); err != nil {
		if errors.Is(err, domain.ErrSubnetNotFound) {
			s.sendError(w, "Subnet not found in whitelist", http.StatusNotFound)
		} else {
			s.sendError(w, fmt.Sprintf("Failed to remove from whitelist: %v", err), http.StatusInternalServerError)
		}
		return
	}

	s.sendJSON(w, map[string]string{"message": "Subnet removed from whitelist successfully"}, http.StatusOK)
}

func (s *Server) authHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ip := r.URL.Query().Get("ip")
	if ip == "" {
		s.sendError(w, "IP parameter is required", http.StatusBadRequest)
		return
	}

	response, err := s.app.CheckIPAccess(ip)
	if err != nil {
		if isValidationError(err) {
			s.sendError(w, err.Error(), http.StatusBadRequest)
			return
		}

		s.logger.Error(fmt.Sprintf("IP check failed for %s: %v", ip, err))
		s.sendError(w, fmt.Sprintf("IP check failed: %v", err), http.StatusInternalServerError)
		return
	}

	s.sendJSON(w, response, http.StatusOK)
}

func isValidationError(err error) bool {
	errorMsg := err.Error()
	return strings.Contains(errorMsg, "invalid IP address") ||
		strings.Contains(errorMsg, "only IPv4 addresses are supported")
}

func (s *Server) convertSubnetsToResponse(subnets []domain.Subnet) SubnetsListResponse {
	response := SubnetsListResponse{
		Subnets: make([]SubnetResponse, len(subnets)),
		Count:   len(subnets),
	}

	for i, subnet := range subnets {
		response.Subnets[i] = SubnetResponse{
			ListType: subnet.ListType,
			CIDR:     subnet.CIDR,
		}
	}

	return response
}

func (s *Server) sendJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to encode response: %v", err))
	}
}

func (s *Server) sendError(w http.ResponseWriter, message string, statusCode int) {
	s.sendJSON(w, ErrorResponse{Error: message}, statusCode)
}
