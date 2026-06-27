# ADR-0011: Rezerwacja miejsc i idempotencja płatności

## Status

Zaakceptowano

| Pole | Treść |
| :---- | :---- |
| **Decyzja** | Utrzymujemy przepływ dwuetapowy: rezerwacja miejsca z blokadą czasową, a następnie płatność z obowiązkowym Idempotency-Key. |
| **Kontekst** | Największe ryzyko domenowe to podwójna sprzedaż miejsca i wielokrotne obciążenie tej samej rezerwacji przy ponowieniach po stronie klienta. |
| **Alternatywy** | 1) Jednoetapowy zakup bez blokady; 2) Brak idempotencji i poleganie na kliencie; 3) Model wyłącznie event-sourcing na starcie projektu. |
| **Uzasadnienie** | Dwuetapowy model stabilizuje współbieżność przy wyborze miejsca, a idempotencja zabezpiecza płatności i upraszcza obsługę błędów sieciowych. |
| **Kompromisy** | 1) Więcej statusów i przejść stanów; 2) Potrzeba czyszczenia wygasłych rezerwacji; 3) Większa odpowiedzialność za spójność semantyki kodów błędów. |




