# ADR-0001: Architektura rdzenia TicketStream

## Status

Zaakceptowano

| Pole | Treść |
| :---- | :---- |
| **Decyzja** | Przyjmujemy architekturę transakcyjną z warstwą czasu rzeczywistego i przetwarzaniem asynchronicznym: blokada w Redisie dla rezerwacji, weryfikacja stanu miejsca w transakcji PostgreSQL, idempotentna płatność, wzorzec transactional outbox, konsument RabbitMQ z mechanizmem ponowień i DLQ, WebSocket z numeracją sekwencyjną oraz proces czyszczący wygasłe rezerwacje. |
| **Kontekst** | System sprzedaży biletów musi działać pod ruchem skokowym i jednocześnie eliminować podwójną sprzedaż miejsca. Potrzebne są też szybkie aktualizacje stanu miejsc dla wielu klientów równocześnie. |
| **Alternatywy** | 1) Tylko SQL bez blokady w Redisie (większe okno kolizji); 2) Tylko blokada w pamięci procesu API (brak spójności między instancjami); 3) Cykliczne odpytywanie HTTP zamiast WebSocket (większy ruch i gorsze doświadczenie użytkownika); 4) Zadania poboczne synchronicznie w API (większe opóźnienia). |
| **Uzasadnienie** | Wybrany układ rozdziela odpowiedzialności: krytyczna ścieżka transakcyjna pozostaje krótka i bezpieczna, a zadania poboczne są odporne na awarie dzięki outboxowi i kolejce. Warstwa czasu rzeczywistego utrzymuje spójny obraz zajętości miejsc po stronie warstwy klienckiej. |
| **Kompromisy** | 1) Większa złożoność systemu i operacji; 2) Więcej komponentów infrastruktury do monitorowania; 3) Konieczność konsekwentnej idempotencji i testów integracyjnych. |




