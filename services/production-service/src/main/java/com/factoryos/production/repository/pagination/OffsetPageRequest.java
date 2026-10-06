package com.factoryos.production.repository.pagination;

import java.util.List;
import java.util.Objects;

import org.springframework.data.domain.Pageable;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Sort;

import com.factoryos.common.error.CommonErrorCode;
import com.factoryos.common.error.DomainException;
import com.factoryos.common.error.ErrorKey;

public record OffsetPageRequest(int page, int pageSize, List<SortCriteria> sortCriteriaList) {

    private static final int DEFAULT_PAGE_SIZE = 20;
    private static final int MAX_PAGE_SIZE = 100;

    public OffsetPageRequest {
        Objects.requireNonNull(sortCriteriaList, "sortCriteriaList must not be null");
        if (sortCriteriaList.isEmpty()) {
            throw new DomainException(CommonErrorCode.MALFORMED_REQUEST,
                    "sortCriteriaList must not be empty");
        }
        if (page < 0) {
            throw new DomainException(CommonErrorCode.PAGE_SIZE_OUT_OF_RANGE,
                    "page must be >= 0, got %d".formatted(page),
                    List.of(new DomainException.FieldError(ErrorKey.PAGE, "must be >= 0",
                            String.valueOf(page), List.of())));
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

    public static OffsetPageRequest of(int rawPage, int rawPageSize,
            List<SortCriteria> sortCriteriaList) {
        int pageSize = (rawPageSize == 0) ? DEFAULT_PAGE_SIZE : rawPageSize;
        return new OffsetPageRequest(rawPage, pageSize, sortCriteriaList);
    }

    public static OffsetPageRequest ofApiPage(int apiPage, int rawPageSize,
            List<SortCriteria> sortCriteriaList) {
        if (apiPage < 1) {
            throw new DomainException(CommonErrorCode.PAGE_SIZE_OUT_OF_RANGE,
                    "page must be >= 1, got %d".formatted(apiPage),
                    List.of(new DomainException.FieldError(ErrorKey.PAGE, "must be >= 1",
                            String.valueOf(apiPage), List.of())));
        }
        int pageSize = (rawPageSize == 0) ? DEFAULT_PAGE_SIZE : rawPageSize;
        return new OffsetPageRequest(apiPage - 1, pageSize, sortCriteriaList);
    }

    public Pageable toPageable() {
        Sort sort = toSort();
        return PageRequest.of(page, pageSize, sort);
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
