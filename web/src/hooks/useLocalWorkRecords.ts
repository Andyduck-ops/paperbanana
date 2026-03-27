import { useState, useCallback, useEffect } from 'react';

export interface LocalWorkRecord {
  id: string;
  createdAt: string;
  status: string;
  prompt?: string;
  summary?: string;
  mode: 'generate' | 'batch' | 'refine';
  candidateSessionIds?: string[];
}

export interface LocalWorkRecordsState {
  records: LocalWorkRecord[];
  addRecord: (record: LocalWorkRecord) => void;
  removeRecord: (id: string) => void;
  clearRecords: () => void;
}

const LOCAL_WORK_RECORDS_KEY = 'paperbanana-work-records';
const MAX_RECORDS = 24;

function loadRecords(): LocalWorkRecord[] {
  if (typeof window === 'undefined') return [];
  
  try {
    const stored = localStorage.getItem(LOCAL_WORK_RECORDS_KEY);
    if (!stored) return [];
    
    const parsed = JSON.parse(stored) as LocalWorkRecord[];
    // Validate and sort by createdAt descending
    return parsed
      .filter((r): r is LocalWorkRecord => 
        r && typeof r.id === 'string' && typeof r.createdAt === 'string'
      )
      .sort((a, b) => 
        new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
      )
      .slice(0, MAX_RECORDS);
  } catch {
    return [];
  }
}

function saveRecords(records: LocalWorkRecord[]): void {
  if (typeof window === 'undefined') return;
  
  try {
    localStorage.setItem(LOCAL_WORK_RECORDS_KEY, JSON.stringify(records));
  } catch {
    // Ignore storage errors (e.g., quota exceeded)
  }
}

export function useLocalWorkRecords(): LocalWorkRecordsState {
  const [records, setRecords] = useState<LocalWorkRecord[]>([]);
  
  // Load records on mount
  useEffect(() => {
    setRecords(loadRecords());
  }, []);
  
  const addRecord = useCallback((record: LocalWorkRecord) => {
    setRecords((prev) => {
      // Remove duplicate if exists, add new record at front, limit to MAX_RECORDS
      const filtered = prev.filter((r) => r.id !== record.id);
      const next = [record, ...filtered].slice(0, MAX_RECORDS);
      saveRecords(next);
      return next;
    });
  }, []);
  
  const removeRecord = useCallback((id: string) => {
    setRecords((prev) => {
      const next = prev.filter((r) => r.id !== id);
      saveRecords(next);
      return next;
    });
  }, []);
  
  const clearRecords = useCallback(() => {
    setRecords([]);
    saveRecords([]);
  }, []);
  
  return {
    records,
    addRecord,
    removeRecord,
    clearRecords,
  };
}
