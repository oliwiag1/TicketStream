# ADR-0004: Monorepo dla warstwy serwerowej, warstwy klienckiej i infrastruktury

## Status

Zaakceptowano

| Pole | Treść |
| :---- | :---- |
| **Decyzja** | Utrzymujemy jedno monorepo dla kodu aplikacyjnego, konfiguracji wdrożeń i dokumentacji. |
| **Kontekst** | Zmiany w kontrakcie API często wpływają jednocześnie na warstwę kliencką, testy, docker-compose i scenariusz operacyjny prezentacji. |
| **Alternatywy** | 1) Polyrepo (oddzielne repozytoria dla warstwy serwerowej, warstwy klienckiej i infrastruktury); 2) Osobne repo dokumentacyjne; 3) Monorepo tylko dla kodu, bez części wdrożeniowej i dokumentacyjnej. |
| **Uzasadnienie** | Monorepo upraszcza synchronizację zmian przekrojowych i zmniejsza ryzyko rozjazdu wersji między warstwami systemu. Daje też jeden, spójny punkt uruchomienia i przegląd. |
| **Kompromisy** | 1) Szerszy zakres zmian w pojedynczych PR; 2) Potrzeba dyscypliny katalogów i konwencji; 3) Wydłużony czas wykonywania potoku CI przy braku selektywnego uruchamiania zadań. |




