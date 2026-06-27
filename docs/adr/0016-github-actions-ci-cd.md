# ADR-0016: CI/CD na GitHub Actions

## Status

Zaakceptowano

| Pole | Treść |
| :---- | :---- |
| **Decyzja** | Uruchamiamy potok CI/CD w GitHub Actions: lint i testy na push/PR, a budowanie obrazów na push do gałęzi main. |
| **Kontekst** | Potrzebna jest automatyczna kontrola jakości po każdej zmianie oraz szybkie wychwytywanie regresji przed prezentacją i oddaniem projektu. |
| **Alternatywy** | 1) GitLab CI; 2) Ręczne uruchamianie testów lokalnie bez CI; 3) Potok CI tylko backendowy bez warstwy klienckiej. |
| **Uzasadnienie** | Actions integruje się natywnie z repozytorium i ma niski koszt utrzymania, a jednocześnie spełnia formalne wymaganie „lint + test przy każdym pushu”. |
| **Kompromisy** | 1) Dłuższa informacja zwrotna dla drobnych zmian; 2) Trzeba utrzymywać wersje akcji i stabilność przebiegu CI; 3) Budowanie bez publikacji do rejestru obrazów to niepełny proces wydawniczy. |




