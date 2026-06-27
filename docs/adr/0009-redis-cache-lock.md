# ADR-0009: Redis dla pamięci podręcznej i blokad

## Status

Zaakceptowano

| Pole | Treść |
| :---- | :---- |
| **Decyzja** | Redis używamy do pamięci podręcznej listy wydarzeń oraz krótkotrwałych blokad miejsc podczas rezerwacji. |
| **Kontekst** | Lista wydarzeń jest często odczytywana, a wybór miejsc wymaga szybkiej koordynacji konkurujących żądań. |
| **Alternatywy** | 1) Pamięć podręczna HTTP wyłącznie po stronie klienta; 2) Tylko blokady i pamięć podręczna w SQL; 3) Pamięć podręczna w pamięci operacyjnej dla każdej instancji API. |
| **Uzasadnienie** | Redis dobrze obsługuje TTL, operacje atomowe i niskie opóźnienia, co jednocześnie wspiera wydajność odczytu i bezpieczeństwo procesu rezerwacji. |
| **Kompromisy** | 1) Potencjalna niespójność nieaktualnej pamięci podręcznej; 2) Kolejny komponent infrastrukturalny; 3) Konieczność pilnowania semantyki zwalniania blokad i TTL. |




