# ADR-0003: Wybór technologii warstwy klienckiej (TypeScript + React)

## Status

Zaakceptowano

| Pole | Treść |
| :---- | :---- |
| **Decyzja** | Warstwę kliencką budujemy jako SPA w React + TypeScript, z bundlerem Vite. |
| **Kontekst** | UI musi obsłużyć logowanie OIDC, listę wydarzeń, mapę miejsc oraz aktualizacje czasu rzeczywistego statusów miejsc. |
| **Alternatywy** | 1) Vue + TypeScript; 2) Svelte/SvelteKit; 3) SSR w Next.js. |
| **Uzasadnienie** | React + TypeScript daje dojrzały ekosystem i czytelny model komponentów, a Vite skraca czas budowania i uruchamiania lokalnego. Integracja z React Query dobrze wspiera stan asynchroniczny. |
| **Kompromisy** | 1) SPA wymaga większej dyscypliny w obsłudze pamięci podręcznej i sesji; 2) Brak SSR nie pomaga SEO (akceptowalne dla aplikacji transakcyjnej); 3) Dodatkowa złożoność typowania danych API. |




