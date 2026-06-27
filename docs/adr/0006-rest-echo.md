# ADR-0006: REST API na Echo

## Status

Zaakceptowano

| Pole | Treść |
| :---- | :---- |
| **Decyzja** | Warstwę API realizujemy jako REST HTTP na frameworku Echo. |
| **Kontekst** | Warstwa kliencka SPA potrzebuje prostych punktów końcowych do operacji CRUD i przepływu rezerwacja → płatność, z czytelną semantyką kodów HTTP. |
| **Alternatywy** | 1) GraphQL; 2) gRPC; 3) net/http bez frameworka. |
| **Uzasadnienie** | REST jest najbardziej bezpośredni do debugowania i prezentacji, a Echo dostarcza lekkie middleware i routing bez dużego narzutu. |
| **Kompromisy** | 1) Ryzyko nadmiarowego pobierania danych względem GraphQL; 2) Konieczność ręcznego utrzymywania kontraktu punktów końcowych; 3) Mniejsza „sztywność” kontraktu niż w gRPC. |




