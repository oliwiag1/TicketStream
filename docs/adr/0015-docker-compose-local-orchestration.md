# ADR-0015: Docker Compose jako orkiestracja lokalna

## Status

Zaakceptowano

| Pole | Treść |
| :---- | :---- |
| **Decyzja** | Standardem uruchamiania lokalnego i prezentacji jest Docker Compose obejmujący wszystkie usługi systemu. |
| **Kontekst** | Projekt ma wiele zależności uruchomieniowych i musi startować jedną komendą w sposób przewidywalny dla zespołu i prowadzącego. |
| **Alternatywy** | 1) Ręczne uruchamianie każdej usługi osobno; 2) Lokalny Kubernetes jako domyślny tryb dev; 3) Zestaw skryptów bez deklaratywnego compose. |
| **Uzasadnienie** | Docker Compose upraszcza wdrożenie nowych członków zespołu i odtwarzalność środowiska, wspiera kontrole zdrowia i zależności startowe oraz pasuje do wymagań zaliczenia. |
| **Kompromisy** | 1) To nie jest pełna symulacja produkcyjnego orkiestratora; 2) Ograniczone możliwości stopniowego wdrażania i autoskalowania; 3) Konieczność utrzymywania definicji usług w jednym pliku. |




