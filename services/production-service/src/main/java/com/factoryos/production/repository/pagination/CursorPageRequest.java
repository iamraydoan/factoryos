package com.factoryos.production.repository.pagination;

import java.util.List;
import java.util.Objects;

import org.springframework.data.domain.Sort;

import com.factoryos.common.error.CommonErrorCode;
import com.factoryos.common.error.DomainException;
import com.factoryos.common.error.ErrorKey;

public record CursorPageRequest(int pageSize, Cursor cursor, List<SortCriteria> sortCriteriaList) {

    private static final int DEFAULT_PAGE_SIZE = 20;
    private static final int MAX_PAGE_SIZE = 100;

    public CursorPageRequest {
        Objects.requireNonNull(sortCriteriaList, "sortCriteriaList must not be null");
        if (sortCriteriaList.isEmpty()) {
            throw new IllegalArgumentException("sortCriteriaList must not be empty");
        }
        if (pageSize < 1 || pageSize > MAX_PAGE_SIZE) {
            throw new DomainException(CommonErrorCode.PAGE_SIZE_OUT_OF_RANGE,
                    "pageSize must be between 1 and %d, got %d".formatted(MAX_PAGE_SIZE, pageSize),
                    List.of(new DomainException.FieldError(ErrorKey.LIMIT,
                            "must be between 1 and %d".formatted(MAX_PAGE_SIZE),
                            String.valueOf(pageSize),
                            List.of("1..%d".formatted(MAX_PAGE_SIZE)))));
        }
    }

    public static CursorPageRequest of(int rawPageSize, String pageToken,
            List<SortCriteria> sortCriteriaList) {
        // 0 (not specified) → default, negative or > max → error
        int pageSize = (rawPageSize == 0) ? DEFAULT_PAGE_SIZE : rawPageSize;
        // Extract sort keys for cursor decoding
        List<SortKey<?>> sortKeys = sortCriteriaList.stream()
                .<SortKey<?>>map(SortCriteria::key)
                .toList();
        // Decode cursor: null/empty → null (first page), invalid → throws
        // InvalidCursorException
        Cursor cursor = Cursor.decode(pageToken, sortKeys);
        return new CursorPageRequest(pageSize, cursor, sortCriteriaList);
    }

    public int queryLimit() {
        return pageSize + 1;
    }

    public Sort toSort() {
        Sort result = null;
        for (SortCriteria criteria : sortCriteriaList) {
            Sort s = Sort.by(criteria.direction(), criteria.key().fieldName());
            result = (result == null) ? s : result.and(s);
        }
        return result;
    }
}
