import { Injectable } from '@angular/core';

@Injectable({ providedIn: 'root' })
export class MobileStateService {
  selectedEveningId   = '';
  lastCatchUpPlayedDate = '';
}
