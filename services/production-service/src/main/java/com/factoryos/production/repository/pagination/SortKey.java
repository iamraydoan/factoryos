package com.factoryos.production.repository.pagination;

import java.time.Instant;
import java.util.Objects;
import java.util.UUID;
import java.util.function.Function;

public record SortKey<V>(String fieldName, Class<V> type, Function<String, V> valueParser) {
    public SortKey {
        Objects.requireNonNull(fieldName, "fieldName must not be null");
        Objects.requireNonNull(type, "type must not be null");
        Objects.requireNonNull(valueParser, "valueParser must not be null");
    }

    public V parseValue(String value) {
        if (value == null || value.isEmpty()) {
            throw new IllegalArgumentException(
                    "Sort key value for %s cannot be null or empty".formatted(fieldName));
        }
        try {
            return valueParser.apply(value);
        } catch (Exception e) {
            throw new IllegalArgumentException(
                    "Invalid value for sort key '%s': '%s' (%s)".formatted(fieldName, value, e.getMessage()), e);
        }
    }

    public static SortKey<UUID> id() {
        return new SortKey<>("id", UUID.class, UUID::fromString);
    }

    public static SortKey<Instant> ofTimestamp(String fieldName) {
        return new SortKey<>(fieldName, Instant.class, Instant::parse);
    }

    public static SortKey<String> ofString(String fieldName) {
        return new SortKey<>(fieldName, String.class, Function.identity());
    }

    public static <V> SortKey<V> of(String fieldName, Class<V> type, Function<String, V> valueParser) {
        return new SortKey<>(fieldName, type, valueParser);
    }
}
