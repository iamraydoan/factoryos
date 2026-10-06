package com.factoryos.production;

import com.factoryos.common.error.ErrorCategory;
import com.factoryos.common.error.ErrorCode;
import com.factoryos.common.error.ErrorSeverity;

/**
 * Error codes owned by the Production bounded context.
 *
 * <p>
 * Domain codes live with their domain, not in the shared library — a shared code
 * must be resolvable by every service, and these are meaningful only here.
 */
public enum WorkOrderErrorCode implements ErrorCode {
    WORK_ORDER_NOT_FOUND(ErrorCategory.NOT_FOUND,
            "Work order not found", ErrorSeverity.INFO, false),
    WORK_ORDER_INVALID_TRANSITION(ErrorCategory.CONFLICT,
            "Invalid state transition", ErrorSeverity.WARNING, false),
    WORK_ORDER_ALREADY_EXISTS(ErrorCategory.CONFLICT,
            "Work order already exists", ErrorSeverity.WARNING, false),

    /**
     * A lost optimistic lock. It shares the {@code CONFLICT} category with the two
     * codes above, but its gRPC status differs: the category default is
     * FAILED_PRECONDITION (9), while a retryable concurrency loss is ABORTED (10).
     * That override is declared in the gRPC adapter's table (see
     * {@link WorkOrderErrorOverrides}), not by the code's name.
     */
    WORK_ORDER_CONCURRENT_MODIFICATION(ErrorCategory.CONFLICT,
            "Work order was modified concurrently", ErrorSeverity.WARNING, true);

    private final ErrorCategory category;
    private final String title;
    private final ErrorSeverity severity;
    private final boolean retryable;

    private WorkOrderErrorCode(ErrorCategory category, String title, ErrorSeverity severity, boolean retryable) {
        this.category = category;
        this.title = title;
        this.severity = severity;
        this.retryable = retryable;
    }

    @Override
    public ErrorCategory category() {
        return category;
    }

    @Override
    public String title() {
        return title;
    }

    @Override
    public ErrorSeverity severity() {
        return severity;
    }

    @Override
    public boolean retryable() {
        return retryable;
    }
}
