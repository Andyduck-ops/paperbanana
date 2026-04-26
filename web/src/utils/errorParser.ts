export interface FieldError {
  field: string;
  message: string;
}

export function parseValidationErrors(err: unknown): FieldError[] {
  if (!err) return [];

  if (err instanceof Error) {
    try {
      const parsed = JSON.parse(err.message);
      if (Array.isArray(parsed.errors)) {
        return parsed.errors.map((e: { field?: string; message?: string }) => ({
          field: e.field || 'general',
          message: e.message || 'Validation failed',
        }));
      }
      if (Array.isArray(parsed.details)) {
        return parsed.details.map((d: { field?: string; message?: string }) => ({
          field: d.field || 'general',
          message: d.message || 'Validation failed',
        }));
      }
      if (parsed.field && parsed.message) {
        return [{ field: parsed.field, message: parsed.message }];
      }
    } catch {
      return [{ field: 'general', message: err.message }];
    }
    return [{ field: 'general', message: err.message }];
  }

  if (typeof err === 'object' && err !== null) {
    const obj = err as Record<string, unknown>;
    if (Array.isArray(obj.errors)) {
      return obj.errors.map((e) => ({
        field: (e as { field?: string }).field || 'general',
        message: (e as { message?: string }).message || 'Validation failed',
      }));
    }
    if (Array.isArray(obj.details)) {
      return obj.details.map((d) => ({
        field: (d as { field?: string }).field || 'general',
        message: (d as { message?: string }).message || 'Validation failed',
      }));
    }
    if (obj.field && obj.message) {
      return [{ field: String(obj.field), message: String(obj.message) }];
    }
  }

  return [{ field: 'general', message: String(err) }];
}
