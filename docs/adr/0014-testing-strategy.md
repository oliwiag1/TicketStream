# ADR-0014: Strategia testów wielowarstwowych

## Status

Zaakceptowano

| Pole | Treść |
| :---- | :---- |
| **Uzasadnienie** | Podejście warstwowe pozwala szybko wychwycić regresje lokalne i równocześnie walidować rzeczywiste przepływy między DB, Redis, RabbitMQ i WebSocket. |
| **Decyzja** | Łączymy testy jednostkowe, integracyjne, scenariuszowe (end-to-end) i obciążeniowe, aby pokryć logikę krytyczną i zachowanie pod obciążeniem. |
| **Kontekst** | Największe ryzyko systemu leży we współbieżności, idempotencji i spójności między komponentami, czego nie pokryje pojedynczy typ testów. |
| **Alternatywy** | 1) Tylko testy jednostkowe; 2) Tylko testy scenariuszowe; 3) Brak testów obciążeniowych. |
| **Kompromisy** | 1) Dłuższy czas działania potoku CI; 2) Większy nakład na utrzymanie danych testowych i konfiguracji testów integracyjnych; 3) Potrzeba stabilizacji testów zależnych od infrastruktury. |




