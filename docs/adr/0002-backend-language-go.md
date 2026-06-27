# ADR-0002: Wybór języka warstwy serwerowej (Go)

## Status

Zaakceptowano

| Pole | Treść |
| :---- | :---- |
| **Decyzja** | Warstwę serwerową implementujemy w Go 1.22. |
| **Kontekst** | Kluczowe punkty końcowe rezerwacji i płatności muszą być przewidywalne pod obciążeniem, a wdrożenie ma być proste w kontenerach. |
| **Alternatywy** | 1) Node.js/TypeScript; 2) Java/Kotlin; 3) Python. |
| **Uzasadnienie** | Go zapewnia niski narzut środowiska uruchomieniowego, prostą współbieżność oraz szybkie, statyczne binarki. To dobrze pasuje do usług API i procesu roboczego uruchamianych w Dockerze. |
| **Kompromisy** | 1) Mniej „magii frameworkowej”, więcej kodu jawnego; 2) Konieczność ręcznego mapowania części DTO i walidacji; 3) Mniejszy wybór bibliotek w niektórych niszach niż w ekosystemach dynamicznych. |




