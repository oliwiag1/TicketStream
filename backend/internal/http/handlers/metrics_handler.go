package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"ticketstream/backend/internal/observability"

	"github.com/labstack/echo/v4"
)

type MetricsHandler struct{}

func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{}
}

func (h *MetricsHandler) JSON(c echo.Context) error {
	return c.JSON(http.StatusOK, observability.GetSnapshot())
}

func (h *MetricsHandler) Dashboard(c echo.Context) error {
	snapshot := observability.GetSnapshot()
	var rows strings.Builder
	for _, route := range snapshot.Routes {
		rows.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%d</td><td>%d</td><td>%.2f</td><td>%.2f</td></tr>", route.Route, route.Count, route.Errors, route.P95Ms, route.P99Ms))
	}

	html := fmt.Sprintf(`<!doctype html>
<html>
<head>
  <meta charset="utf-8" />
  <title>TicketStream Ops Dashboard</title>
  <style>
    body { font-family: Arial, sans-serif; margin: 20px; background: #f4f6fb; color: #172033; }
    .cards { display: grid; grid-template-columns: repeat(4, minmax(120px, 1fr)); gap: 12px; margin-bottom: 16px; }
    .card { background: #fff; border: 1px solid #d7dee8; border-radius: 8px; padding: 12px; }
    .label { font-size: 12px; color: #475569; text-transform: uppercase; }
    .value { font-size: 22px; font-weight: bold; }
    table { width: 100%%; border-collapse: collapse; background: #fff; border: 1px solid #d7dee8; }
    th, td { border: 1px solid #d7dee8; padding: 8px; font-size: 14px; text-align: left; }
    th { background: #eef2f8; }
  </style>
</head>
<body>
  <h1>TicketStream Ops Dashboard</h1>
  <p>Generated at: %s</p>
  <section class="cards">
    <article class="card"><div class="label">Requests</div><div class="value">%d</div></article>
    <article class="card"><div class="label">Error Rate</div><div class="value">%.2f%%%%</div></article>
    <article class="card"><div class="label">Queue Lag</div><div class="value">%d</div></article>
    <article class="card"><div class="label">Lock Contention</div><div class="value">%d</div></article>
  </section>
  <table>
    <thead><tr><th>Route</th><th>Count</th><th>Errors</th><th>P95 (ms)</th><th>P99 (ms)</th></tr></thead>
    <tbody>%s</tbody>
  </table>
</body>
</html>`, snapshot.GeneratedAtUTC, snapshot.TotalRequests, snapshot.ErrorRate, snapshot.QueueLag, snapshot.LockContention, rows.String())

	return c.HTML(http.StatusOK, html)
}
