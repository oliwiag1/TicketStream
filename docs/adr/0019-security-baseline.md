# ADR-0019: Bezpieczeństwo - poziom bazowy

## Status

Zaakceptowano

| Pole | Treść |
| :---- | :---- |
| **Decyzja** | Wprowadzamy bazowy zestaw zabezpieczeń: kontrolę CSRF dla mutacji, ograniczanie tempa żądań z użyciem Redis, nagłówki bezpieczeństwa oraz obowiązkową weryfikację bearer JWT na punktach końcowych chronionych. |
| **Kontekst** | Aplikacja działa przez przeglądarkę i obsługuje operacje mutujące, więc musi ograniczać nadużycia i ryzyka typowe dla warstwy web. |
| **Alternatywy** | 1) Tylko uwierzytelnianie bez ograniczania tempa żądań i CSRF; 2) Oparcie się wyłącznie na zewnętrznym WAF; 3) Brak polityk nagłówków bezpieczeństwa. |
| **Uzasadnienie** | Taki bazowy poziom daje realną ochronę przy akceptowalnym koszcie implementacji i dobrze wpisuje się w skalę projektu zaliczeniowego. |
| **Kompromisy** | 1) Możliwe fałszywie dodatnie blokady przy złej konfiguracji listy dozwolonych źródeł; 2) Trudniejsza diagnostyka błędów 403/429; 3) Konieczność utrzymania spójnych polityk między API i nginx. |




