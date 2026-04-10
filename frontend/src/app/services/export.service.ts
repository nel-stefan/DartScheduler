import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';

@Injectable({ providedIn: 'root' })
export class ExportService {
  private http = inject(HttpClient);
  private base = environment.apiBaseUrl;

  downloadExcel(): void {
    window.open(`${this.base}/export/excel`, '_blank');
  }

  downloadPdf(): void {
    window.open(`${this.base}/export/pdf`, '_blank');
  }

  downloadStandingsPdf(scheduleId?: string, listId?: string | null): void {
    const params = new URLSearchParams();
    if (scheduleId) params.set('scheduleId', scheduleId);
    if (listId) params.set('listId', listId);
    const query = params.toString();
    window.open(`${this.base}/stats/pdf${query ? '?' + query : ''}`, '_blank');
  }
}
