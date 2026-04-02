import { create } from 'zustand';

export interface LogEntry {
  id: number;
  timestamp: string;
  level: string;
  source_app: string;
  message: string;
}

interface LogState {
  logs: LogEntry[];
  availableServices: string[];
  isLive: boolean;
  filterLevel: string | null;
  searchQueries: string[];
  selectedService: string | null;
  dateFrom: string;
  dateTo: string;
  setLogs: (logs: LogEntry[]) => void;
  setAvailableServices: (s: string[]) => void;
  addLiveLog: (log: LogEntry) => void;
  toggleLive: () => void;
  setFilterLevel: (level: string | null) => void;
  setSearchQueries: (q: string[]) => void;
  addSearchQuery: (q: string) => void;
  removeSearchQuery: (q: string) => void;
  setSelectedService: (s: string | null) => void;
  setDateFrom: (d: string) => void;
  setDateTo: (d: string) => void;
}

export const useLogStore = create<LogState>((set) => ({
  logs: [],
  availableServices: [],
  isLive: true,
  filterLevel: null,
  searchQueries: [],
  selectedService: null,
  dateFrom: '',
  dateTo: '',
  setLogs: (logs) => set((state) => {
    const services = new Set(state.availableServices);
    logs.forEach(l => {
      if (l.source_app && l.source_app !== 'unknown') services.add(l.source_app);
    });
    return { logs, availableServices: Array.from(services).sort() };
  }),
  setAvailableServices: (newServices) => set((state) => {
    const services = new Set([...state.availableServices, ...newServices]);
    return { availableServices: Array.from(services).sort() };
  }),
  addLiveLog: (newLog) => set((state) => {
    const services = new Set(state.availableServices);
    if (newLog.source_app && newLog.source_app !== 'unknown') {
      services.add(newLog.source_app);
    }
    const nextServices = Array.from(services).sort();

    if (state.filterLevel && newLog.level !== state.filterLevel) {
       return { availableServices: nextServices };
    }
    if (state.selectedService && newLog.source_app !== state.selectedService) {
       return { availableServices: nextServices };
    }
    if (state.dateFrom && newLog.timestamp < state.dateFrom.replace('T', ' ')) {
       return { availableServices: nextServices };
    }
    if (state.dateTo) {
        const toLimit = state.dateTo.includes(':') && state.dateTo.split(':').length === 2 
            ? state.dateTo.replace('T', ' ') + ':59' 
            : state.dateTo.replace('T', ' ');
        if (newLog.timestamp > toLimit) {
            return { availableServices: nextServices };
        }
    }
    
    // Limit store size to 50,000 logs in memory to prevent browser crash
    const updated = [newLog, ...state.logs];
    if (updated.length > 50000) {
      updated.pop();
    }
    return { logs: updated, availableServices: nextServices };
  }),
  toggleLive: () => set((state) => ({ isLive: !state.isLive })),
  setFilterLevel: (level) => set({ filterLevel: level }),
  setSearchQueries: (searchQueries) => set({ searchQueries }),
  addSearchQuery: (q) => set((state) => ({ searchQueries: state.searchQueries.includes(q) ? state.searchQueries : [...state.searchQueries, q] })),
  removeSearchQuery: (q) => set((state) => ({ searchQueries: state.searchQueries.filter(s => s !== q) })),
  setSelectedService: (selectedService) => set({ selectedService }),
  setDateFrom: (dateFrom) => set({ dateFrom }),
  setDateTo: (dateTo) => set({ dateTo }),
}));
