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
}
