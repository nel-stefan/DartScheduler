import { Component, inject, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormBuilder, Validators, AbstractControl } from '@angular/forms';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { Match, Player } from '../models';
import { ScoreService } from '../services/score.service';

function displayName(name: string): string {
  const idx = name.indexOf(', ');
  return idx >= 0 ? `${name.slice(idx + 2)} ${name.slice(0, idx)}` : name;
}

@Component({
  selector: 'app-mobile-score',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, MatSnackBarModule],
  styles: [`
    :host { display: block; min-height: 100dvh; background: #f5f0ee; }

    .top-bar {
      display: flex; align-items: center; gap: 8px;
      background: #4e342e; color: #fff; padding: 12px 16px;
      position: sticky; top: 0; z-index: 10;
    }
    .back-btn {
      background: none; border: none; color: #fff; cursor: pointer;
      font-size: 22px; line-height: 1; padding: 0; display: flex; align-items: center;
    }
    .top-bar h2 { margin: 0; font-size: 16px; font-weight: 500; }

    .match-players {
      background: #5d4037; color: #fff; padding: 10px 16px;
      display: flex; align-items: center; justify-content: space-between; gap: 8px;
    }
    .match-player { font-size: 14px; font-weight: 600; flex: 1; }
    .match-player.b { text-align: right; }
    .match-vs { font-size: 12px; color: rgba(255,255,255,.6); }

    .body { padding: 12px; display: flex; flex-direction: column; gap: 12px; }

    .card {
      background: #fff; border-radius: 10px; padding: 14px;
      box-shadow: 0 1px 3px rgba(0,0,0,.1);
    }
    .card-title {
      font-size: 12px; font-weight: 600; text-transform: uppercase;
      letter-spacing: .5px; color: #6d4c41; margin: 0 0 10px;
    }

    .winner-row {
      display: grid; grid-template-columns: 1fr 1fr; gap: 8px; margin-bottom: 10px;
    }
    .winner-btn {
      padding: 12px 8px; border-radius: 8px; border: 2px solid #e0d5d0;
      background: #faf7f5; font-size: 13px; font-weight: 500; cursor: pointer;
      text-align: center; transition: border-color .15s, background .15s;
      word-break: break-word;
    }
    .winner-btn.selected { border-color: #4e342e; background: #efebe9; }

    .field-row {
      display: flex; align-items: center; gap: 10px;
      margin-bottom: 0;
    }
    .field-label { font-size: 13px; color: #424242; min-width: 70px; }
    .field-input {
      flex: 1; padding: 8px 10px; border: 1.5px solid #d7ccc8; border-radius: 8px;
      font-size: 14px; background: #faf7f5; outline: none;
    }
    .field-input:focus { border-color: #4e342e; }

    .duty-grid {
      display: grid; grid-template-columns: 1fr 1fr; gap: 10px;
    }
    .duty-field label { font-size: 11px; color: #757575; display: block; margin-bottom: 4px; }
    .duty-field input {
      width: 100%; padding: 8px 10px; border: 1.5px solid #d7ccc8; border-radius: 8px;
      font-size: 14px; background: #faf7f5; outline: none; box-sizing: border-box;
    }
    .duty-field input:focus { border-color: #4e342e; }

    .submit-btn {
      width: 100%; padding: 14px; border: none; border-radius: 10px;
      background: #4e342e; color: #fff; font-size: 16px; font-weight: 600;
      cursor: pointer; margin-top: 4px;
    }
    .submit-btn:disabled { background: #bcaaa4; cursor: default; }

    .error { color: #c62828; font-size: 12px; margin-top: 4px; }
  `],
  template: `
    <div class="top-bar">
      <button class="back-btn" (click)="goBack()">&#8592;</button>
      <h2>Score invoeren</h2>
    </div>

    <div class="match-players" *ngIf="match">
      <div class="match-player">{{ nameA }}</div>
      <div class="match-vs">vs</div>
      <div class="match-player b">{{ nameB }}</div>
    </div>

    <form [formGroup]="form" (ngSubmit)="submit()" class="body" *ngIf="match">

      <!-- Leg 1 -->
      <div class="card">
        <p class="card-title">Leg 1</p>
        <div class="winner-row">
          <button type="button" class="winner-btn"
            [class.selected]="form.get('leg1Winner')?.value === match.playerA"
            (click)="setWinner('leg1Winner', match.playerA)">
            {{ nameA }}
          </button>
          <button type="button" class="winner-btn"
            [class.selected]="form.get('leg1Winner')?.value === match.playerB"
            (click)="setWinner('leg1Winner', match.playerB)">
            {{ nameB }}
          </button>
        </div>
        <div class="field-row">
          <span class="field-label">Beurten</span>
          <input class="field-input" type="number" min="1" max="50"
            formControlName="leg1Turns" placeholder="bijv. 15">
        </div>
      </div>

      <!-- Leg 2 -->
      <div class="card">
        <p class="card-title">Leg 2</p>
        <div class="winner-row">
          <button type="button" class="winner-btn"
            [class.selected]="form.get('leg2Winner')?.value === match.playerA"
            (click)="setWinner('leg2Winner', match.playerA)">
            {{ nameA }}
          </button>
          <button type="button" class="winner-btn"
            [class.selected]="form.get('leg2Winner')?.value === match.playerB"
            (click)="setWinner('leg2Winner', match.playerB)">
            {{ nameB }}
          </button>
        </div>
        <div class="field-row">
          <span class="field-label">Beurten</span>
          <input class="field-input" type="number" min="1" max="50"
            formControlName="leg2Turns" placeholder="bijv. 15">
        </div>
      </div>

      <!-- Leg 3 -->
      <div class="card">
        <p class="card-title">Leg 3</p>
        <div class="winner-row">
          <button type="button" class="winner-btn"
            [class.selected]="form.get('leg3Winner')?.value === match.playerA"
            (click)="setWinner('leg3Winner', match.playerA)">
            {{ nameA }}
          </button>
          <button type="button" class="winner-btn"
            [class.selected]="form.get('leg3Winner')?.value === match.playerB"
            (click)="setWinner('leg3Winner', match.playerB)">
            {{ nameB }}
          </button>
        </div>
        <div class="field-row">
          <span class="field-label">Beurten</span>
          <input class="field-input" type="number" min="1" max="50"
            formControlName="leg3Turns" placeholder="bijv. 15">
        </div>
      </div>

      <!-- Schrijver / Teller -->
      <div class="card">
        <p class="card-title">Dienst</p>
        <div class="duty-grid">
          <div class="duty-field">
            <label>Schrijver nr.</label>
            <input type="text" formControlName="secretaryNr" placeholder="bijv. 12">
          </div>
          <div class="duty-field">
            <label>Teller nr.</label>
            <input type="text" formControlName="counterNr" placeholder="bijv. 7">
          </div>
        </div>
      </div>

      <p class="error" *ngIf="errorMsg">{{ errorMsg }}</p>

      <button type="submit" class="submit-btn" [disabled]="submitting || !formValid()">
        {{ submitting ? 'Opslaan…' : 'Opslaan' }}
      </button>

    </form>
  `,
})
export class MobileScoreComponent implements OnInit {
  private route        = inject(ActivatedRoute);
  private router       = inject(Router);
  private scoreService = inject(ScoreService);
  private snackBar     = inject(MatSnackBar);
  private fb           = inject(FormBuilder);

  match:     Match  | null = null;
  players:   Player[]      = [];
  eveningId  = '';
  submitting = false;
  errorMsg   = '';

  form = this.fb.group({
    leg1Winner: ['', Validators.required],
    leg1Turns:  [null as number | null, [Validators.required, Validators.min(1)]],
    leg2Winner: ['', Validators.required],
    leg2Turns:  [null as number | null, [Validators.required, Validators.min(1)]],
    leg3Winner: ['', Validators.required],
    leg3Turns:  [null as number | null, [Validators.required, Validators.min(1)]],
    secretaryNr: [''],
    counterNr:   [''],
  });

  get nameA(): string { return this.match ? this.playerName(this.match.playerA) : ''; }
  get nameB(): string { return this.match ? this.playerName(this.match.playerB) : ''; }

  ngOnInit(): void {
    const state = history.state as { match?: Match; eveningId?: string; players?: Player[] };
    if (!state?.match) { this.router.navigate(['/m/avond']); return; }

    this.match     = state.match;
    this.eveningId = state.eveningId ?? '';
    this.players   = state.players  ?? [];

    if (this.match.played) {
      this.form.patchValue({
        leg1Winner:  this.match.leg1Winner,
        leg1Turns:   this.match.leg1Turns,
        leg2Winner:  this.match.leg2Winner,
        leg2Turns:   this.match.leg2Turns,
        leg3Winner:  this.match.leg3Winner,
        leg3Turns:   this.match.leg3Turns,
        secretaryNr: this.match.secretaryNr,
        counterNr:   this.match.counterNr,
      });
    }
  }

  setWinner(ctrl: string, playerId: string): void {
    this.form.get(ctrl)?.setValue(playerId);
  }

  formValid(): boolean {
    const v = this.form.value;
    return !!(v.leg1Winner && v.leg1Turns && v.leg2Winner && v.leg2Turns && v.leg3Winner && v.leg3Turns);
  }

  playerName(id: string): string {
    const p = this.players.find(pl => pl.id === id);
    return p ? displayName(p.name) : id;
  }

  goBack(): void {
    this.router.navigate(['/m/avond']);
  }

  submit(): void {
    if (!this.match || !this.formValid()) return;
    const v = this.form.value;
    this.submitting = true;
    this.errorMsg   = '';

    this.scoreService.submitResult(this.match.id, {
      leg1Winner:            v.leg1Winner!,
      leg1Turns:             Number(v.leg1Turns),
      leg2Winner:            v.leg2Winner!,
      leg2Turns:             Number(v.leg2Turns),
      leg3Winner:            v.leg3Winner!,
      leg3Turns:             Number(v.leg3Turns),
      playerA180s:           0,
      playerB180s:           0,
      playerAHighestFinish:  0,
      playerBHighestFinish:  0,
      reportedBy:            '',
      rescheduleDate:        '',
      secretaryNr:           v.secretaryNr ?? '',
      counterNr:             v.counterNr   ?? '',
    }).subscribe({
      next: () => {
        this.submitting = false;
        this.snackBar.open('Score opgeslagen', '', { duration: 2000 });
        this.router.navigate(['/m/avond']);
      },
      error: () => {
        this.submitting = false;
        this.errorMsg   = 'Fout bij opslaan, probeer opnieuw.';
      },
    });
  }
}
