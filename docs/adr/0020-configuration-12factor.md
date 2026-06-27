# ADR-0020: Konfiguracja w stylu 12-factor

## Status

Zaakceptowano

| Pole | Treść |
| :---- | :---- |
| **Decyzja** | Konfigurację uruchomieniową trzymamy poza kodem, głównie w zmiennych środowiskowych i plikach env dla środowiska lokalnego. |
| **Kontekst** | Te same binarki mają działać lokalnie, w CI i w profilach wdrożeniowych z różnymi adresami usług i limitami. |
| **Alternatywy** | 1) Wbudowanie konfiguracji w kodzie; 2) Rozproszone pliki konfiguracyjne bez jednego modelu; 3) Konfiguracja wyłącznie przez sekrety bez lokalnych wartości zastępczych. |
| **Uzasadnienie** | Model zmiennych środowiskowych zwiększa przenośność i upraszcza automatyzację wdrożeń. Pozwala zmieniać parametry działania bez rekompilacji aplikacji. |
| **Kompromisy** | 1) Ryzyko błędnej konfiguracji środowiska; 2) Potrzeba dobrej dokumentacji wartości domyślnych; 3) Mniejsza kontrola typów niż w rozbudowanych frameworkach walidacji konfiguracji. |




