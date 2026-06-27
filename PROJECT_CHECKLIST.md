# TicketStream - Co zostalo do zrobienia, aby zakonczyc projekt

## 1. Cel dokumentu

Ten dokument zawiera tylko rzeczy, ktore trzeba jeszcze wykonac, aby uznac projekt za skonczony.

## 2. Definition of Done (koniec projektu)

Projekt uznajemy za zakonczony, gdy:

1. Dziala pelny flow: login -> lista wydarzen -> wybor miejsca -> rezerwacja -> platnosc -> bilet.
2. Nie ma podwojnej sprzedazy miejsca przy rownoleglych probach zakupu.
3. Frontend dostaje realtime update statusu miejsca przez WebSocket.
4. Platnosc ma idempotencje (powtorzone zadanie nie tworzy duplikatu).
5. Worker przetwarza zadania asynchroniczne z retry i DLQ.
6. Testy krytyczne (integration + load) przechodza.
7. Dokumentacja i instrukcja uruchomienia sa aktualne.

## 3. Priorytet P0 (must have)

### 3.1 Auth i sesja

- [x] Stabilna konfiguracja Keycloak realm/client dla srodowisk dev i test.
- [x] Frontend OIDC Authorization Code + PKCE.
- [x] Backendowa walidacja access tokenu przez issuer/JWKS.
- [x] Endpoint /auth/me i mapowanie roli z tokenu.

Kryterium akceptacji:

1. Endpointy chronione zwracaja 401 bez waznego bearer tokenu.
2. Front poprawnie odswieza token przez klienta OIDC.

### 3.2 Reservation core (anty-oversell)

- [x] Redis lock z TTL 10 min dla wybranego miejsca.
- [x] Finalizacja zakupu z SELECT FOR UPDATE w transakcji.
- [x] Zwolnienie locka po anulowaniu lub wygasnieciu.
- [x] Poprawna odpowiedz 409 dla konfliktu miejsca.

Kryterium akceptacji:

1. Dwie rownolegle proby zakupu tego samego miejsca: tylko jedna konczy sie sukcesem.

### 3.3 Payments i idempotencja

- [x] Wymagany naglowek Idempotency-Key w POST /pay.
- [x] Tabela idempotency i zwrot poprzedniego wyniku dla duplikatu.
- [x] Zapis sukcesu platnosci i zmiany statusu miejsca w jednej logice transakcyjnej.

Kryterium akceptacji:

1. Powtorzony POST /pay z tym samym kluczem nie wykonuje platnosci drugi raz.

### 3.4 Outbox + Worker

- [x] Tworzenie outbox event po sukcesie platnosci.
- [x] Publisher outbox -> RabbitMQ.
- [x] Worker PDF/email z retry.
- [x] DLQ dla bledow trwalych.

Kryterium akceptacji:

1. Zdarzenie po platnosci dociera do workera i jest przetworzone.
2. Blad workera nie powoduje utraty zadania.

### 3.5 Realtime WebSocket

- [x] Realny websocket upgrade i obsluga polaczen.
- [x] Broadcast seat_status_changed po lock/sold/release.
- [x] sequence_number i ochrona przed out-of-order.
- [x] Resync snapshot po reconnect.

Kryterium akceptacji:

1. Dwa klienty widza ta sama zmiane statusu miejsca praktycznie natychmiast.

### 3.6 Frontend MVP flow

- [x] Login view + obsluga sesji.
- [x] Lista wydarzen z API.
- [x] Mapa miejsc i akcja rezerwacji.
- [x] Przejscie do platnosci i obsluga komunikatow konfliktu.
- [x] Integracja websocket update.

Kryterium akceptacji:

1. Da sie wykonac pelny flow uzytkownika bez recznego "patchowania" danych.

## 4. Priorytet P1 (powinno byc przed oddaniem)

### 4.1 Cache i wydajnosc

- [x] Cache GET /events w Redis.
- [x] Inwalidacja cache po POST /events.

### 4.2 Rate limiting i bezpieczenstwo

- [x] Redis-based rate limiting dla rezerwacji i auth.
- [x] CORS allowlist i podstawowe security headers.
- [x] CSRF ochrona dla endpointow mutujacych.

### 4.3 Observability

- [x] Metryki P95/P99, error rate, queue lag, lock contention.
- [x] Logi strukturalne z request_id/trace_id.
- [x] Dashboard minimum (np. Grafana albo prosty panel metryk).

### 4.4 Testy

- [x] Unit testy logiki domenowej.
- [x] Integration test dla rownoleglych rezerwacji.
- [x] Integration test dla idempotencji platnosci.
- [x] Smoke test websocket.
- [x] Load test (minimum scenariusz flash sale).

## 5. Priorytet P2 (mile widziane / production stretch)

- [x] PostgreSQL replica + failover plan.
- [x] Redis Sentinel/Cluster plan.
- [x] Kubernetes/ECS deployment profile.
- [x] Backup/restore test runbook.

## 6. Kolejnosc realizacji (rekomendowana)

1. Auth i sesja.
2. Reservation core (lock + FOR UPDATE).
3. Payments + idempotencja.
4. Outbox + worker.
5. Realtime websocket.
6. Frontend pelny flow.
7. Testy, obserwowalnosc, hardening.

## 7. Blokery i zaleznosci miedzy zadaniami

1. Frontend flow platnosci zalezy od gotowego POST /pay.
2. Realtime frontend zalezy od gotowego websocket backendu.
3. Testy obciazeniowe maja sens dopiero po implementacji lock + FOR UPDATE.
4. Outbox i worker zalezy od finalnej semantyki platnosci.

## 8. Minimalny plan na domkniecie projektu (2 sprinty)

### Sprint 1

- [x] Auth + sesja
- [x] Events/seats API + cache
- [x] Reservation core (lock + transakcja)
- [x] Frontend: lista wydarzen + rezerwacja

### Sprint 2

- [x] Payments + idempotencja
- [x] Outbox + worker + DLQ
- [x] WebSocket realtime
- [x] Testy integration/load
- [ ] Finalne poprawki i demo

## 9. Checklista przed oddaniem

- [ ] Demo przechodzi od A do Z bez restartu i bez "manual fix".
- [x] Scenariusz konfliktu miejsca jest pokazany i dziala poprawnie.
- [x] Scenariusz duplikatu platnosci jest pokazany i dziala poprawnie.
- [x] README zawiera aktualne komendy uruchomienia.
- [x] Architektura i ADR sa zgodne z tym, co jest w kodzie.

## 10. Weryfikacja z przegladu kodu (2026-06-27)

Status realny po wdrozeniach:

1. Dodano automatyczne zwalnianie wygasnietych rezerwacji (sweeper) z aktualizacja statusu miejsca i broadcastem realtime.
2. Wdrozono Redis-based rate limiting dla auth i reserve.
3. Wdrozono security headers i CSRF protection (origin/referer) dla endpointow mutujacych.
4. Wdrozono observability: metryki P95/P99, error rate, queue lag, lock contention, endpoint /metrics i dashboard /ops/dashboard.
5. Dodano structured request logging z request_id i trace_id.
6. Dodano scenariusz load test (flash sale) oraz dokumentacje README, ADR, profile K8s/ECS i runbook backup/restore.
7. Docker Compose utwardzony healthcheckami i warunkami depends_on; potwierdzony cold start bez recznego restartu API/workera.
