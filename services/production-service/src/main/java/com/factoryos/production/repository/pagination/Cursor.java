package com.factoryos.production.repository.pagination;

import java.util.ArrayList;
import java.util.Base64;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Objects;

import tools.jackson.core.type.TypeReference;
import tools.jackson.databind.ObjectMapper;

/**
 * Encodes and decodes opaque, versioned, URL-safe cursor tokens for keyset
 * pagination.
 *
 * <p>
 * Token format: {@code v1.<base64url(JSON)>} where the JSON payload contains
 * {@code keys} (field names) and {@code vals} (string-encoded values).
 *
 * <p>
 * This class is immutable and thread-safe.
 */
public class Cursor {
    private static final String VERSION_PREFIX = "v1.";
    private static final ObjectMapper MAPPER = new ObjectMapper();
    private static final TypeReference<Map<String, List<String>>> MAP_TYPE = new TypeReference<>() {
    };

    private final List<String> keyNames;
    private final List<Object> values;

    private Cursor(List<String> keyNames, List<Object> values) {
        this.keyNames = List.copyOf(keyNames);
        this.values = List.copyOf(values);
    }

    public static Cursor of(List<? extends SortKey<?>> sortKeys, List<Object> values) {
        Objects.requireNonNull(sortKeys, "sortKeys must not be null");
        Objects.requireNonNull(values, "values must not be null");
        if (sortKeys.isEmpty()) {
            throw new IllegalArgumentException("sortKeys must not be empty");
        }
        if (sortKeys.size() != values.size()) {
            throw new IllegalArgumentException("sortKeys and values must have the same size");
        }
        List<String> names = sortKeys.stream().map(SortKey::fieldName).toList();
        return new Cursor(names, values);
    }

    public List<String> keyNames() {
        return keyNames;
    }

    public List<Object> values() {
        return values;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (!(o instanceof Cursor other)) return false;
        return keyNames.equals(other.keyNames) && values.equals(other.values);
    }

    @Override
    public int hashCode() {
        return Objects.hash(keyNames, values);
    }

    public String encode() {
        try {
            Map<String, Object> payload = new LinkedHashMap<>();
            payload.put("keys", keyNames);
            payload.put("vals", values.stream().map(Object::toString).toList());

            byte[] json = MAPPER.writeValueAsBytes(payload);
            return VERSION_PREFIX + Base64.getUrlEncoder().withoutPadding().encodeToString(json);
        } catch (Exception e) {
            throw new IllegalStateException("Failed to encode cursor", e);
        }
    }

    public static Cursor decode(String token, List<? extends SortKey<?>> sortKeys) {
        if (token == null || token.isEmpty()) {
            return null;
        }
        if (!token.startsWith(VERSION_PREFIX)) {
            throw new InvalidCursorException("Invalid cursor token: missing version prefix");
        }

        try {
            String base64 = token.substring(VERSION_PREFIX.length());
            byte[] json = Base64.getUrlDecoder().decode(base64);
            Map<String, List<String>> payload = MAPPER.readValue(json, MAP_TYPE);

            List<String> actualKeys = payload.get("keys");
            List<String> actualVals = payload.get("vals");

            if (actualKeys == null || actualVals == null) {
                throw new InvalidCursorException("Invalid cursor token: missing keys or vals");
            }
            if (actualKeys.size() != actualVals.size()) {
                throw new InvalidCursorException("Invalid cursor token: keys and vals size mismatch");
            }

            // Validate keys match expected sort keys
            List<String> expectedKeys = sortKeys.stream().map(SortKey::fieldName).toList();
            if (!actualKeys.equals(expectedKeys)) {
                throw new InvalidCursorException(
                        "Cursor keys mismatch: expected %s but got %s".formatted(expectedKeys, actualKeys));
            }

            // Parse each value using the corresponding SortKey
            List<Object> values = new ArrayList<>();
            for (int i = 0; i < sortKeys.size(); i++) {
                values.add(sortKeys.get(i).parseValue(actualVals.get(i)));
            }

            return new Cursor(actualKeys, values);
        } catch (InvalidCursorException e) {
            throw e;
        } catch (Exception e) {
            throw new InvalidCursorException("Invalid cursor token: " + e.getMessage(), e);
        }
    }
}
