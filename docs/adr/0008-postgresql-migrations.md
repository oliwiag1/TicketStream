# ADR-0008: PostgreSQL i migracje SQL

## Status

Zaakceptowano

| Pole | Treść |
| :---- | :---- |
| **Decyzja** | Dane transakcyjne przechowujemy w PostgreSQL, a zmiany schematu realizujemy migracjami SQL. |
| **Kontekst** | Model domeny ma relacje i operacje wymagające transakcyjności oraz blokad wierszy przy współbieżności. |
| **Alternatywy** | 1) Baza dokumentowa; 2) SQLite; 3) Migracje ORM bez jawnego SQL. |
| **Uzasadnienie** | PostgreSQL zapewnia ACID, blokowanie na poziomie wiersza i dojrzały ekosystem operacyjny. Jawne migracje zwiększają kontrolę i audytowalność zmian. |
| **Kompromisy** | 1) Większa odpowiedzialność za SQL; 2) Konieczność dbania o kompatybilność migracji; 3) Więcej pracy przy zmianach schematu. |




