import { Component, inject, OnInit, OnDestroy, Inject, signal, computed } from '@angular/core';
import { Subscription } from 'rxjs';

import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatDialog, MatDialogModule, MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { MatSnackBar, MatSnackBarModule } from '@angular/material/snack-bar';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatCardModule } from '@angular/material/card';
import { MatTableModule } from '@angular/material/table';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatSelectModule } from '@angular/material/select';
import { MatChipsModule } from '@angular/material/chips';
import { MatDividerModule } from '@angular/material/divider';
import { PlayerService } from '../../services/player.service';
import { SeasonService } from '../../services/season.service';
import { ScheduleService } from '../../services/schedule.service';
import { Player, PlayerList } from '../../models';

// --- Edit dialog ---

@Component({
  selector: 'app-player-edit-dialog',
  imports: [ReactiveFormsModule, MatDialogModule, MatButtonModule, MatFormFieldModule, MatInputModule, MatSelectModule],
  template: `
    <h2 mat-dialog-title>Speler bewerken</h2>
    <mat-dialog-content>
      <form
        [formGroup]="form"
        style="display:grid;grid-template-columns:1fr 1fr;gap:8px 16px;min-width:480px;padding-top:8px"
      >
        <mat-form-field><mat-label>Nr</mat-label><input matInput formControlName="nr" /></mat-form-field>
        <mat-form-field><mat-label>Naam</mat-label><input matInput formControlName="name" /></mat-form-field>
        <mat-form-field><mat-label>E-mail</mat-label><input matInput formControlName="email" /></mat-form-field>
        <mat-form-field><mat-label>Sponsor</mat-label><input matInput formControlName="sponsor" /></mat-form-field>
        <mat-form-field><mat-label>Adres</mat-label><input matInput formControlName="address" /></mat-form-field>
        <mat-form-field><mat-label>Postcode</mat-label><input matInput formControlName="postalCode" /></mat-form-field>
        <mat-form-field><mat-label>Woonplaats</mat-label><input matInput formControlName="city" /></mat-form-field>
        <mat-form-field><mat-label>Telefoon</mat-label><input matInput formControlName="phone" /></mat-form-field>
        <mat-form-field><mat-label>Mobiel</mat-label><input matInput formControlName="mobile" /></mat-form-field>
        <mat-form-field
          ><mat-label>Lid sinds</mat-label><input matInput formControlName="memberSince"
        /></mat-form-field>
        <mat-form-field>
          <mat-label>Klasse</mat-label>
          <mat-select formControlName="class">
            <mat-option value="">—</mat-option>
            <mat-option value="1">Klasse 1</mat-option>
            <mat-option value="2">Klasse 2</mat-option>
          </mat-select>
        </mat-form-field>
      </form>
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button mat-dialog-close>Annuleren</button>
      <button mat-raised-button color="primary" [disabled]="form.invalid" (click)="submit()">Opslaan</button>
    </mat-dialog-actions>
  `,
})
export class PlayerEditDialogComponent {
  private dialogRef = inject(MatDialogRef<PlayerEditDialogComponent>);
  fb = inject(FormBuilder);

  constructor(@Inject(MAT_DIALOG_DATA) public data: Player) {}

  form = this.fb.group({
    nr: [this.data.nr],
    name: [this.data.name, Validators.required],
    email: [this.data.email],
    sponsor: [this.data.sponsor],
    address: [this.data.address],
    postalCode: [this.data.postalCode],
    city: [this.data.city],
    phone: [this.data.phone],
    mobile: [this.data.mobile],
    memberSince: [this.data.memberSince],
    class: [this.data.class],
  });

  submit(): void {
    if (this.form.valid) {
      this.dialogRef.close({ ...this.data, ...this.form.value } as Player);
    }
  }
}

// --- Buddy dialog ---

@Component({
  selector: 'app-buddy-dialog',
  imports: [MatDialogModule, MatButtonModule, MatCheckboxModule, MatDividerModule],
  template: `
    <h2 mat-dialog-title>Voorkeur speelpartners voor {{ data.player.name }}</h2>
    <mat-dialog-content style="max-height:420px;overflow-y:auto;min-width:320px">
      <p style="color:#666;font-size:13px;margin-top:0">
        Geselecteerde spelers worden bij voorkeur op dezelfde avond ingepland.
      </p>
      <mat-divider style="margin-bottom:12px"></mat-divider>
      @for (p of data.others; track p) {
        <div style="padding:4px 0">
          <mat-checkbox [checked]="selected.has(p.id)" (change)="toggle(p.id)">
            <span style="margin-left:4px">
              <strong>{{ p.nr ? '#' + p.nr + ' ' : '' }}</strong
              >{{ p.name }}
            </span>
          </mat-checkbox>
        </div>
      }
    </mat-dialog-content>
    <mat-dialog-actions align="end">
      <button mat-button mat-dialog-close>Annuleren</button>
      <button mat-raised-button color="primary" (click)="submit()">Opslaan</button>
    </mat-dialog-actions>
  `,
})
export class BuddyDialogComponent {
  private dialogRef = inject(MatDialogRef<BuddyDialogComponent>);

  constructor(@Inject(MAT_DIALOG_DATA) public data: { player: Player; others: Player[]; currentBuddyIds: string[] }) {}

  selected = new Set<string>(this.data.currentBuddyIds);

  toggle(id: string): void {
    if (this.selected.has(id)) this.selected.delete(id);
    else this.selected.add(id);
  }

  submit(): void {
    this.dialogRef.close(Array.from(this.selected));
  }
}

// --- Players page ---

@Component({
  selector: 'app-spelers',
  imports: [
    MatSnackBarModule,
    MatButtonModule,
    MatIconModule,
    MatCardModule,
    MatTableModule,
    MatDialogModule,
    MatChipsModule,
    MatCheckboxModule,
    MatSelectModule,
    MatFormFieldModule,
  ],
  styles: [
    `
      table {
        width: 100%;
      }
      .actions-cell {
        text-align: right;
        white-space: nowrap;
      }
      .buddy-chip {
        font-size: 11px;
      }
      .batch-bar {
        display: flex;
        align-items: center;
        gap: 12px;
        flex-wrap: wrap;
        background: #e3f2fd;
        border-radius: 6px;
        padding: 8px 16px;
        margin-bottom: 12px;
      }
    `,
  ],
  template: `
    <div style="margin-bottom:16px;display:flex;align-items:center;gap:16px;flex-wrap:wrap">
      @if (playerLists().length > 1) {
        <mat-form-field subscriptSizing="dynamic">
          <mat-label>Ledenlijst</mat-label>
          <mat-select [value]="selectedListId()" (selectionChange)="selectedListId.set($event.value)">
            @for (list of playerLists(); track list.id) {
              <mat-option [value]="list.id">{{ list.name }}</mat-option>
            }
          </mat-select>
        </mat-form-field>
      }
      <button mat-stroked-button [disabled]="players().length === 0" (click)="printClassList()">
        <mat-icon>print</mat-icon> Klasseindeling afdrukken
      </button>
    </div>

    @if (filteredPlayers().length > 0) {
      <mat-card>
        <mat-card-header>
          <mat-card-title>Spelers ({{ filteredPlayers().length }})</mat-card-title>
        </mat-card-header>
        <mat-card-content>
          <!-- Batch action bar -->
          @if (selection.size > 0) {
            <div class="batch-bar">
              <span style="font-weight:500">{{ selection.size }} geselecteerd</span>
              <mat-form-field style="min-width:140px" subscriptSizing="dynamic">
                <mat-label>Klasse instellen</mat-label>
                <mat-select [value]="batchClass()" (valueChange)="batchClass.set($event)">
                  <mat-option value="">— geen —</mat-option>
                  <mat-option value="1">Klasse 1</mat-option>
                  <mat-option value="2">Klasse 2</mat-option>
                </mat-select>
              </mat-form-field>
              <button mat-raised-button color="primary" (click)="applyBatchClass()">Toepassen</button>
              <button mat-button (click)="selection.clear()">Deselecteer</button>
            </div>
          }
          <table mat-table [dataSource]="filteredPlayers()">
            <!-- Checkbox column -->
            <ng-container matColumnDef="select">
              <th mat-header-cell *matHeaderCellDef style="width:40px">
                <mat-checkbox
                  [checked]="allSelected()"
                  [indeterminate]="selection.size > 0 && !allSelected()"
                  (change)="toggleAll($event.checked)"
                >
                </mat-checkbox>
              </th>
              <td mat-cell *matCellDef="let p" style="width:40px">
                <mat-checkbox [checked]="selection.has(p.id)" (change)="toggleOne(p.id)"></mat-checkbox>
              </td>
            </ng-container>
            <ng-container matColumnDef="nr">
              <th mat-header-cell *matHeaderCellDef style="width:48px">#</th>
              <td mat-cell *matCellDef="let p">{{ p.nr }}</td>
            </ng-container>
            <ng-container matColumnDef="name">
              <th mat-header-cell *matHeaderCellDef>Naam</th>
              <td mat-cell *matCellDef="let p">
                <strong>{{ p.name }}</strong>
              </td>
            </ng-container>
            <ng-container matColumnDef="city">
              <th mat-header-cell *matHeaderCellDef>Woonplaats</th>
              <td mat-cell *matCellDef="let p">{{ p.city }}</td>
            </ng-container>
            <ng-container matColumnDef="class">
              <th mat-header-cell *matHeaderCellDef style="width:110px">Klasse</th>
              <td mat-cell *matCellDef="let p">
                <mat-select
                  [value]="p.class"
                  (selectionChange)="updateClass(p, $event.value)"
                  style="font-size:14px"
                  panelWidth=""
                >
                  <mat-option value="">—</mat-option>
                  <mat-option value="1">Klasse 1</mat-option>
                  <mat-option value="2">Klasse 2</mat-option>
                </mat-select>
              </td>
            </ng-container>
            <ng-container matColumnDef="buddies">
              <th mat-header-cell *matHeaderCellDef>Speelpartners</th>
              <td mat-cell *matCellDef="let p">
                @if (buddyMap()[p.id] && buddyMap()[p.id].length) {
                  <mat-chip-set>
                    @for (bid of buddyMap()[p.id]; track bid) {
                      <mat-chip class="buddy-chip" disableRipple>
                        {{ playerName(bid) }}
                      </mat-chip>
                    }
                  </mat-chip-set>
                }
                @if (!buddyMap()[p.id] || !buddyMap()[p.id].length) {
                  <span style="color:#bbb">—</span>
                }
              </td>
            </ng-container>
            <ng-container matColumnDef="actions">
              <th mat-header-cell *matHeaderCellDef></th>
              <td mat-cell *matCellDef="let p" class="actions-cell">
                <button mat-icon-button color="primary" (click)="openEdit(p)" matTooltip="Bewerken">
                  <mat-icon>edit</mat-icon>
                </button>
                <button mat-icon-button color="accent" (click)="openBuddies(p)" matTooltip="Speelpartners instellen">
                  <mat-icon>group</mat-icon>
                </button>
                <button mat-icon-button color="warn" (click)="deletePlayer(p)" matTooltip="Lid verwijderen">
                  <mat-icon>delete</mat-icon>
                </button>
              </td>
            </ng-container>
            <tr mat-header-row *matHeaderRowDef="cols"></tr>
            <tr mat-row *matRowDef="let row; columns: cols"></tr>
          </table>
        </mat-card-content>
      </mat-card>
    }

    @if (filteredPlayers().length === 0) {
      <p style="color:#888;text-align:center;padding:32px 0">
        {{ players().length === 0 ? 'Nog geen spelers geïmporteerd.' : 'Geen spelers in deze lijst.' }}
      </p>
    }
  `,
})
export class SpelersComponent implements OnInit, OnDestroy {
  private playerService   = inject(PlayerService);
  private seasonService   = inject(SeasonService);
  private scheduleService = inject(ScheduleService);
  private snackBar        = inject(MatSnackBar);
  private dialog          = inject(MatDialog);

  players         = signal<Player[]>([]);
  playerLists     = signal<PlayerList[]>([]);
  selectedListId  = signal<string | null>(null);
  filteredPlayers = computed(() => {
    const id = this.selectedListId();
    return id ? this.players().filter((p) => p.listId === id) : this.players();
  });
  buddyMap      = signal<Record<string, string[]>>({});
  selection     = new Set<string>();
  batchClass    = signal('');
  private readonly classListNote =
    'Als je een avond niet kunt gooien dien je je uiterlijk donderdagavond voor 8u af te ' +
    'melden bij je tegenstander en bij de wedstrijdleider: Stefan Marchal 06-24201115';
  cols = ['select', 'nr', 'name', 'class', 'city', 'buddies', 'actions'];
  private sub = new Subscription();

  ngOnInit(): void {
    this.loadPlayers();
    this.playerService.getPlayerLists().subscribe({
      next: (lists) => this.playerLists.set(lists),
      error: () => {},
    });
    this.sub.add(
      this.seasonService.effectivePlayerListId$.subscribe((id) => this.selectedListId.set(id)),
    );
  }

  ngOnDestroy(): void {
    this.sub.unsubscribe();
  }

  loadPlayers(): void {
    this.playerService.list().subscribe({
      next: (ps) => {
        this.players.set(ps);
        this.loadAllBuddies();
      },
      error: () => {},
    });
  }

  loadAllBuddies(): void {
    this.buddyMap.set({});
    for (const p of this.players()) {
      this.playerService.getBuddies(p.id).subscribe({
        next: (ids) => {
          this.buddyMap.set({ ...this.buddyMap(), [p.id]: ids });
        },
        error: () => {},
      });
    }
  }

  playerName(id: string): string {
    return this.players().find((p) => p.id === id)?.name ?? id.slice(0, 8);
  }

  allSelected(): boolean {
    return this.filteredPlayers().length > 0 && this.selection.size === this.filteredPlayers().length;
  }

  toggleAll(checked: boolean): void {
    if (checked) this.filteredPlayers().forEach((p) => this.selection.add(p.id));
    else this.selection.clear();
  }

  toggleOne(id: string): void {
    if (this.selection.has(id)) this.selection.delete(id);
    else this.selection.add(id);
  }

  updateClass(player: Player, cls: string): void {
    this.playerService.update({ ...player, class: cls }).subscribe({
      next: (p) => {
        this.players.set(this.players().map((x) => (x.id === p.id ? p : x)));
      },
      error: (err) => this.snackBar.open(`Fout: ${err.message}`, 'Sluiten', { duration: 5000 }),
    });
  }

  applyBatchClass(): void {
    const ids = Array.from(this.selection);
    let done = 0;
    for (const id of ids) {
      const player = this.players().find((p) => p.id === id);
      if (!player) continue;
      this.playerService.update({ ...player, class: this.batchClass() }).subscribe({
        next: (p) => {
          this.players.set(this.players().map((x) => (x.id === p.id ? p : x)));
          if (++done === ids.length) {
            this.snackBar.open(`Klasse bijgewerkt voor ${done} spelers`, 'OK', { duration: 2000 });
            this.selection.clear();
          }
        },
        error: (err) => this.snackBar.open(`Fout: ${err.message}`, 'Sluiten', { duration: 5000 }),
      });
    }
  }

  deletePlayer(player: Player): void {
    if (!confirm(`Lid "${player.name}" verwijderen inclusief alle wedstrijden?`)) return;
    this.playerService.delete(player.id).subscribe({
      next: () => {
        this.players.set(this.players().filter((p) => p.id !== player.id));
        this.selection.delete(player.id);
        this.snackBar.open('Lid verwijderd', 'OK', { duration: 2000 });
      },
      error: (err) => this.snackBar.open(`Fout: ${err.message}`, 'Sluiten', { duration: 5000 }),
    });
  }

  openEdit(player: Player): void {
    const ref = this.dialog.open(PlayerEditDialogComponent, { data: player });
    ref.afterClosed().subscribe((updated: Player | undefined) => {
      if (!updated) return;
      this.playerService.update(updated).subscribe({
        next: (p) => {
          this.players.set(this.players().map((x) => (x.id === p.id ? p : x)));
          this.snackBar.open('Speler opgeslagen', 'OK', { duration: 2000 });
        },
        error: (err) => this.snackBar.open(`Fout: ${err.message}`, 'Sluiten', { duration: 5000 }),
      });
    });
  }

  openBuddies(player: Player): void {
    const currentBuddyIds = this.buddyMap()[player.id] ?? [];
    const ref = this.dialog.open(BuddyDialogComponent, {
      data: {
        player,
        others: this.filteredPlayers().filter((p) => p.id !== player.id),
        currentBuddyIds,
      },
    });
    ref.afterClosed().subscribe((buddyIds: string[] | undefined) => {
      if (buddyIds === undefined) return;
      this.playerService.setBuddies(player.id, buddyIds).subscribe({
        next: () => {
          this.buddyMap.set({ ...this.buddyMap(), [player.id]: buddyIds });
          this.snackBar.open('Speelpartners opgeslagen', 'OK', { duration: 2000 });
        },
        error: (err) => this.snackBar.open(`Fout: ${err.message}`, 'Sluiten', { duration: 5000 }),
      });
    });
  }

  printClassList(): void {
    const note = this.classListNote.trim();
    const selectedId = this.seasonService.selectedId$.value;
    const seasons = this.seasonService.seasons$.value;
    const season = seasons.find((s) => s.id === selectedId);
    const compName = season?.competitionName ?? '';
    const seasonLabel = season?.season ?? '';

    const sorted = [...this.players()].sort(
      (a, b) => (parseInt(a.nr) || 9999) - (parseInt(b.nr) || 9999),
    );
    const byClass = new Map<string, Player[]>();
    for (const p of sorted) {
      const cls = p.class || '—';
      byClass.set(cls, [...(byClass.get(cls) ?? []), p]);
    }
    const classes = [...byClass.keys()].sort();

    const listHtml = classes
      .map((cls) => {
        const rows = byClass
          .get(cls)!
          .map((p) => `<tr><td class="nr">${p.nr}</td><td>${p.name}</td></tr>`)
          .join('');
        return `<div class="class-block">
          <h3>${cls}</h3>
          <table><thead><tr><th class="nr">Nr.</th><th>Naam</th></tr></thead>
          <tbody>${rows}</tbody></table>
        </div>`;
      })
      .join('');

    const html = `<!DOCTYPE html><html><head><meta charset="utf-8">
      <title>Klasseindeling${seasonLabel ? ' ' + seasonLabel : ''}</title>
      <style>
        body { font-family: Arial, sans-serif; font-size: 13px; padding: 20px; }
        h1 { font-size: 17px; text-align: center; margin: 0 0 6px; text-transform: uppercase; letter-spacing: .5px; }
        h2 { font-size: 13px; text-align: center; color: #555; margin: 0 0 20px; font-weight: normal; }
        .lists { display: flex; gap: 40px; flex-wrap: wrap; }
        .class-block { flex: 1; min-width: 160px; }
        h3 { font-size: 14px; text-transform: uppercase; letter-spacing: .4px; margin: 0 0 6px;
             border-bottom: 2px solid #333; padding-bottom: 3px; }
        table { border-collapse: collapse; width: 100%; }
        th { font-weight: 600; text-align: left; padding: 3px 6px; border-bottom: 1px solid #bbb; font-size: 12px; }
        td { padding: 3px 6px; border-bottom: 1px solid #eee; }
        .nr { width: 40px; color: #555; }
        .note { margin-top: 24px; font-size: 12px; color: #333; border-top: 1px solid #ccc; padding-top: 10px; }
        @media print { @page { size: A4 portrait; margin: 15mm; } }
      </style></head><body>
      <h1>Klasse indeling en naam + nummers${seasonLabel ? ' ' + seasonLabel : ''}</h1>
      ${compName ? `<h2>${compName}</h2>` : ''}
      <div class="lists">${listHtml}</div>
      ${note ? `<div class="note">${note.replace(/\n/g, '<br>')}</div>` : ''}
      <script>window.onload = () => { window.print(); }<\/script>
      </body></html>`;

    const w = window.open('', '_blank');
    if (w) {
      w.document.write(html);
      w.document.close();
    }
  }
}
