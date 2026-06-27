# ADR-0005: Struktura folderów projektu

## Status

Zaakceptowano

| Pole | Treść |
| :---- | :---- |
| **Decyzja** | Przyjmujemy podział katalogów: backend/, frontend/, deploy/, docs/, a w katalogu backend/ układ cmd/internal/pkg. |
| **Kontekst** | Projekt ma być łatwy do oceny i utrzymania, a granice odpowiedzialności muszą być czytelne dla zespołu i prowadzącego. |
| **Alternatywy** | 1) Struktura płaska bez wyraźnych granic; 2) Sam układ domenowy bez cmd/internal/pkg; 3) Dokumentacja poza repozytorium. |
| **Uzasadnienie** | Taki podział porządkuje kod i upraszcza nawigację: punkty wejścia są oddzielone od logiki biznesowej, a konfiguracja wdrożeniowa i dokumentacja nie mieszają się z kodem uruchomieniowym. |
| **Kompromisy** | 1) Więcej katalogów na start; 2) Wymagana konsekwencja przy umieszczaniu nowych plików; 3) Konieczność utrzymywania opisów struktury w README. |




