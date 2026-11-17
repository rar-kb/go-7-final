package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

func TestCafeCount(t *testing.T) {
	city := "moscow"

	handler := http.HandlerFunc(mainHandle)
	want100 := len(cafeList[city])
	if want100 > 100 {
		want100 = 100
	}

	requests := []struct {
		count int
		want  int
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{100, want100},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/cafe?city=%s&count=%d", city, v.count), nil)

		handler.ServeHTTP(response, req)

		require.Equal(t, http.StatusOK, response.Code)

		body := strings.TrimSpace(response.Body.String())

		var got int
		if body != "" {
			got = len(strings.Split(body, ","))
		}

		require.Equal(t, v.want, got, "count=%d: body=%q", v.count, body)
	}
}

func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		search    string
		wantCount int
	}{
		{"фасоль", 0},
		{"кофе", 2},
		{"вилка", 1},
	}

	for _, tc := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest(
			http.MethodGet,
			"/cafe?city=moscow&search="+url.QueryEscape(tc.search),
			nil,
		)
		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)

		body := strings.TrimSpace(response.Body.String())

		var cafes []string
		if body != "" {
			cafes = strings.Split(body, ",")
		}

		assert.Len(t, cafes, tc.wantCount,
			"search=%q: body=%q", tc.search, body)

		for _, cafe := range cafes {
			cafe = strings.TrimSpace(cafe)
			assert.True(t, strings.Contains(strings.ToLower(cafe), strings.ToLower(tc.search)), "кафе %q не содержит %q", cafe, tc.search)
		}
	}
}
