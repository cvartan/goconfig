Простой менеджер для работы с конфигурациями, использующие конфигурационные файлы в JSON формате (по умолчанию, но можно реализовать ридер, использующий другие форматы, например, YAML).

Поддерживает также получение конфигурационной информации из переменных окружения:
- автоматическое переопределение значения параметры конфигурации по наличию переменной окружения в специальном формате
- задание ссылки на переменную окружения как значения параметра окружения
