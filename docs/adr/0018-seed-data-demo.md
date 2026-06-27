# ADR-0018: Dane inicjalne dla prezentacji i testów manualnych

## Status

Zaakceptowano

| Pole | Treść |
| :---- | :---- |
| **Decyzja** | Dane startowe (wydarzenia i miejsca) utrzymujemy jako część migracji SQL. |
| **Kontekst** | Prezentacja i testy manualne muszą być powtarzalne bez ręcznego tworzenia danych po każdym starcie środowiska. |
| **Alternatywy** | 1) Osobny skrypt inicjalizacyjny uruchamiany ręcznie; 2) Tworzenie danych przez punkty końcowe administracyjne; 3) Brak danych inicjalnych i przygotowanie danych „na żywo”. |
| **Uzasadnienie** | Dane inicjalne w migracjach uruchamiają się automatycznie i gwarantują spójny stan startowy, co redukuje ryzyko błędów podczas prezentacji. |
| **Kompromisy** | 1) Potrzeba pilnowania idempotencji danych inicjalnych; 2) Konieczność aktualizacji danych inicjalnych przy zmianach modelu; 3) Dodatkowy kod SQL do utrzymania. |




