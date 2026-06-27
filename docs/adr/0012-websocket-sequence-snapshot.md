# ADR-0012: Komunikacja czasu rzeczywistego przez WebSocket (migawka stanu + sekwencja)

## Status

Zaakceptowano

| Pole | Treść |
| :---- | :---- |
| **Decyzja** | Statusy miejsc propagujemy przez WebSocket, z pełną migawką stanu po połączeniu i numerem sekwencji dla zdarzeń inkrementalnych. |
| **Kontekst** | Wiele klientów jednocześnie obserwuje te same wydarzenia i wymaga spójnych, niskolatencyjnych aktualizacji bez agresywnego pollingu HTTP. |
| **Alternatywy** | 1) cykliczne odpytywanie HTTP co kilka sekund; 2) Server-Sent Events; 3) Brak czasu rzeczywistego i ręczne odświeżanie. |
| **Uzasadnienie** | WebSocket zmniejsza opóźnienia i liczbę żądań. Migawka stanu i numer sekwencyjny pomagają odzyskać spójny stan po ponownym połączeniu i wykryć luki. |
| **Kompromisy** | 1) Większa złożoność zarządzania połączeniami; 2) Trudniejsze testowanie scenariuszy ponownego łączenia; 3) Wymóg ostrożnej konfiguracji warstwy pośredniej dla upgrade do WS. |




