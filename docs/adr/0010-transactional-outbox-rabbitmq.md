# ADR-0010: Transactional outbox i RabbitMQ

## Status

Zaakceptowano

| Pole | Treść |
| :---- | :---- |
| **Decyzja** | Po płatności zapisujemy zdarzenie do outboxu w tej samej transakcji, a publikację i przetwarzanie wykonujemy asynchronicznie przez RabbitMQ i konsumenta. |
| **Kontekst** | Trzeba uniknąć utraty zdarzenia między COMMIT transakcji a wysyłką do brokera oraz odciążyć ścieżkę API od zadań pobocznych. |
| **Alternatywy** | 1) Bezpośrednia publikacja z handlera płatności bez outboxu; 2) Synchroniczne wykonywanie wszystkiego w API; 3) Cykliczne odpytywanie bez brokera. |
| **Uzasadnienie** | Outbox rozwiązuje problem podwójnego zapisu (dual-write), a RabbitMQ zapewnia trasowanie, ponawianie i DLQ. Dzięki temu system jest odporniejszy na awarie i restart komponentów. |
| **Kompromisy** | 1) Większa złożoność operacyjna; 2) Spójność ostateczna zamiast efektu natychmiastowego wszędzie; 3) Wymagana idempotencja konsumenta. |




