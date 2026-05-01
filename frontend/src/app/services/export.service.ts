import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';

@Injectable({ providedIn: 'root' })
export class ExportService {
  private http = inject(HttpClient);
  private base = environment.apiBaseUrl;

  downloadExcel(): void {
    this.downloadBlob(`${this.base}/export/excel`, 'schema.xlsx');
  }

  downloadPdf(): void {
    this.openBlob(`${this.base}/export/pdf`);
  }

  downloadStandingsPdf(scheduleId?: string, listId?: string | null): void {
    const params = new URLSearchParams();
    if (scheduleId) params.set('scheduleId', scheduleId);
    if (listId) params.set('listId', listId);
    const query = params.toString();
    this.openBlob(`${this.base}/stats/pdf${query ? '?' + query : ''}`);
  }

  downloadEveningExcel(eveningId: string): void {
    this.downloadBlob(`${this.base}/export/evening/${eveningId}/excel`, `avond-${eveningId}.xlsx`);
  }

  downloadEveningPdf(eveningId: string, date: string): void {
    this.downloadBlob(`${this.base}/export/evening/${eveningId}/pdf`, `wedstrijdformulier_${date.slice(0, 10)}.pdf`);
  }

  openEveningPrint(eveningId: string): void {
    this.openBlob(`${this.base}/export/evening/${eveningId}/print`);
  }

  private downloadBlob(url: string, filename: string): void {
    this.http.get(url, { responseType: 'blob' }).subscribe((blob) => {
      const objectUrl = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = objectUrl;
      a.download = filename;
      a.click();
      URL.revokeObjectURL(objectUrl);
    });
  }

  private openBlob(url: string): void {
    this.http.get(url, { responseType: 'blob' }).subscribe((blob) => {
      const objectUrl = URL.createObjectURL(blob);
      window.open(objectUrl, '_blank');
      setTimeout(() => URL.revokeObjectURL(objectUrl), 30000);
    });
  }
}
