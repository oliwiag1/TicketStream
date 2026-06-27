# ADR-0017: OpenAPI jako kontrakt API

## Status

Zaakceptowano

| Pole | Treść |
| :---- | :---- |
| **Decyzja** | Utrzymujemy formalną dokumentację interfejsu w specyfikacji OpenAPI 3.0.3. |
| **Kontekst** | API jest rozwijane równolegle z frontendem i podlega ocenie, więc kontrakt powinien być jednoznaczny, czytelny i łatwy do weryfikacji. |
| **Alternatywy** | 1) Postman collection bez formalnego schematu; 2) Dokumentacja tylko tekstowa w README; 3) Brak dokumentacji kontraktu. |
| **Uzasadnienie** | OpenAPI to standard wspierany przez narzędzia walidujące i wizualizatory, co ułatwia przegląd techniczny oraz synchronizację warstwy serwerowej i warstwy klienckiej. |
| **Kompromisy** | 1) Potrzeba aktualizacji specyfikacji przy zmianach punktów końcowych; 2) Ryzyko dryfu, jeśli brak dyscypliny zespołu; 3) Dodatkowy koszt utrzymaniowy dokumentacji. |




