# Ссылки

## Параллельные тесты с БД
- [Pytest-xdist: как распараллелить тесты без конфликтов в базе (KTS, Хабр)](https://habr.com/ru/companies/kts/articles/912840/) — схема Postgres на воркер, номер БД Redis на воркер, ×4,7 в примере, ×2,5 на проекте.
- [Как распараллелить тесты с базой данных (Хабр)](https://habr.com/ru/articles/778580/) — копия БД на процесс для 800+ легаси-тестов, разбор альтернатив, ×2 локально и ×4 в CI.
- [pytest-xdist: распараллеливание тестов в pytest (Medium)](https://medium.com/@rudikqa/pytest-xdist-%D1%80%D0%B0%D1%81%D0%BF%D0%B0%D1%80%D0%B0%D0%BB%D0%BB%D0%B5%D0%BB%D0%B8%D0%B2%D0%B0%D0%BD%D0%B8%D0%B5-%D1%82%D0%B5%D1%81%D1%82%D0%BE%D0%B2-%D0%B2-pytest-e647a4d2e1b1)
- [How to parallelize integration tests (r/devops)](https://www.reddit.com/r/devops/comments/10c9z54/how_to_parallelize_integration_tests/)
- [Как сделать тестовые окружения воспроизводимыми и не забить диск (Astana Hub)](https://astanahub.com/ru/blog/kak-sdelat-testovye-okruzheniia-vosproizvodimymi-i-ne-zabit-disk) — окружение на задачу, уборка ресурсов, бюджет диска при параллельности.

## Kafka
- [Isolate Kafka consumer tests (The New Stack)](https://thenewstack.io/isolate-kafka-consumer-tests/) — метка изоляции в каждом сообщении.
- [Signadot's Kafka routing key trick (DEV)](https://dev.to/leobaniak/signadots-kafka-routing-key-trick-one-broker-many-parallel-tests-2jpg)
- [Isolation in testing with Kafka (DZone)](https://dzone.com/articles/isolation-in-testing-with-kafka)
- [Kafka testing (Conduktor)](https://www.conduktor.io/blog/kafka-testing.md)
- [Kafka test automation, correlation id (Opcito)](https://www.opcito.com/blogs/kafka-test-automation-spring-boot-event-driven-framework)
- [Testcontainers Kafka consumer groups (qaskills)](https://qaskills.sh/blog/testcontainers-kafka-consumer-groups)

## Инструменты и документация
- [Testo](https://github.com/ozontech/testo) — фреймворк поверх `testing.T`, [пример параллельности](https://github.com/ozontech/testo/blob/main/examples/05_parallel/main_test.go), [статья на Хабре](https://habr.com/companies/ozontech/articles/1062036).
- [Testcontainers for Go](https://golang.testcontainers.org/)
- [Postgres: CREATE DATABASE … TEMPLATE](https://www.postgresql.org/docs/current/manage-ag-templatedbs.html)
- [Postgres: advisory locks](https://www.postgresql.org/docs/current/explicit-locking.html#ADVISORY-LOCKS)
- [Postgres: pg_stat_activity и pg_blocking_pids](https://www.postgresql.org/docs/current/monitoring-stats.html#MONITORING-PG-STAT-ACTIVITY-VIEW)
- [go test flags (-p, -parallel, -count, -shuffle)](https://pkg.go.dev/cmd/go#hdr-Testing_flags)
- [Redis SELECT (логические БД)](https://redis.io/docs/latest/commands/select/)
