# ADR-0013: Obserwowalność - poziom bazowy

## Status

Zaakceptowano

| Pole | Treść |
| :---- | :---- |
| **Decyzja** | Przyjmujemy bazowy poziom obserwowalności: logi strukturalne, punkty końcowe health/live/ready oraz metryki prezentowane na uproszczonym pulpicie. |
| **Kontekst** | Architektura wielousługowa wymaga szybkiej diagnozy awarii i degradacji bez pełnej platformy APM na starcie projektu. |
| **Alternatywy** | 1) Tylko logi tekstowe; 2) Pełne rozproszone śledzenie od pierwszej wersji; 3) Zewnętrzna platforma obserwowalności jako obowiązkowa zależność. |
| **Uzasadnienie** | Ten poziom telemetryczny daje wysoki stosunek wartości do kosztu: można szybko sprawdzić stan systemu, opóźnienia i błędy oraz skorelować żądania. |
| **Kompromisy** | 1) Brak pełnego śledzenia przepływu żądania przez cały system; 2) Metryki bez długoterminowej retencji; 3) Uproszczony pulpit względem narzędzi klasy enterprise. |




