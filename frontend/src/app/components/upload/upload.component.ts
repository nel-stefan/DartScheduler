import { Component, inject, OnInit, Inject, ViewChild, ElementRef, ChangeDetectorRef } from '@angular/core';
import { CommonModule } from '@angular/common';
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
import { Player } from '../../models';

// --- Edit dialog ---

@Component({
  selector: 'app-player-edit-dialog',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, MatDialogModule, MatButtonModule,
            MatFormFieldModule, MatInputModule, MatSelectModule],
  template: `
    <h2 mat-dialog-title>Speler bewerken</h2>
    <mat-dialog-content>
      <form [formGroup]="form" style="display:grid;grid-template-columns:1fr 1fr;gap:8px 16px;min-width:480px;padding-top:8px">
        <mat-form-field><mat-label>Nr</mat-label><input matInput formControlName="nr"></mat-form-field>
        <mat-form-field><mat-label>Naam</mat-label><input matInput formControlName="name"></mat-form-field>
        <mat-form-field><mat-label>E-mail</mat-label><input matInput formControlName="email"></mat-form-field>
        <mat-form-field><mat-label>Sponsor</mat-label><input matInput formControlName="sponsor"></mat-form-field>
        <mat-form-field><mat-label>Adres</mat-label><input matInput formControlName="address"></mat-form-field>
        <mat-form-field><mat-label>Postcode</mat-label><input matInput formControlName="postalCode"></mat-form-field>
        <mat-form-field><mat-label>Woonplaats</mat-label><input matInput formControlName="city"></mat-form-field>
        <mat-form-field><mat-label>Telefoon</mat-label><input matInput formControlName="phone"></mat-form-field>
        <mat-form-field><mat-label>Mobiel</mat-label><input matInput formControlName="mobile"></mat-form-field>
        <mat-form-field><mat-label>Lid sinds</mat-label><input matInput formControlName="memberSince"></mat-form-field>
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
    nr:          [this.data.nr],
    name:        [this.data.name, Validators.required],
    email:       [this.data.email],
    sponsor:     [this.data.sponsor],
    address:     [this.data.address],
    postalCode:  [this.data.postalCode],
    city:        [this.data.city],
    phone:       [this.data.phone],
    mobile:      [this.data.mobile],
    memberSince: [this.data.memberSince],
    class:       [this.data.class],
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
  standalone: true,
  imports: [CommonModule, MatDialogModule, MatButtonModule, MatCheckboxModule, MatDividerModule],
  template: `
    <h2 mat-dialog-title>Voorkeur speelpartners voor {{ data.player.name }}</h2>
    <mat-dialog-content style="max-height:420px;overflow-y:auto;min-width:320px">
      <p style="color:#666;font-size:13px;margin-top:0">
        Geselecteerde spelers worden bij voorkeur op dezelfde avond ingepland.
      </p>
      <mat-divider style="margin-bottom:12px"></mat-divider>
      <div *ngFor="let p of data.others" style="padding:4px 0">
        <mat-checkbox
          [checked]="selected.has(p.id)"
          (change)="toggle(p.id)">
          <span style="margin-left:4px">
            <strong>{{ p.nr ? ('#' + p.nr + ' ') : '' }}</strong>{{ p.name }}
          </span>
        </mat-checkbox>
      </div>
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
  selector: 'app-upload',
  standalone: true,
  imports: [
    CommonModule, MatSnackBarModule, MatButtonModule, MatIconModule,
    MatCardModule, MatTableModule, MatDialogModule, MatChipsModule,
    MatCheckboxModule, MatSelectModule, MatFormFieldModule,
  ],
  styles: [`
    .import-row { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
    table { width: 100%; }
    .actions-cell { text-align: right; white-space: nowrap; }
    .buddy-chip { font-size: 11px; }
    .batch-bar {
      display: flex; align-items: center; gap: 12px; flex-wrap: wrap;
      background: #e3f2fd; border-radius: 6px; padding: 8px 16px; margin-bottom: 12px;
    }
  `],
  template: `
    <mat-card style="margin-bottom:24px">
      <mat-card-header>
        <mat-card-title>Spelers importeren</mat-card-title>
        <mat-card-subtitle>Upload een Excel-bestand met de ledenlijst</mat-card-subtitle>
      </mat-card-header>
      <mat-card-content style="padding-top:16px">
        <div class="import-row">
          <input #fileInput type="file" accept=".xlsx,.xls" hidden (change)="onFileSelected($event)">
          <button mat-raised-button color="primary" (click)="fileInput.click()">
            <mat-icon>upload_file</mat-icon> Excel-bestand kiezen
          </button>
          <span *ngIf="selectedFile" style="color:#555">{{ selectedFile.name }}</span>
          <button mat-raised-button color="accent" [disabled]="!selectedFile || loading" (click)="upload()">
            {{ loading ? 'Bezig…' : 'Importeren' }}
          </button>
        </div>
      </mat-card-content>
    </mat-card>

    <mat-card *ngIf="players.length > 0">
      <mat-card-header>
        <mat-card-title>Spelers ({{ players.length }})</mat-card-title>
      </mat-card-header>
      <mat-card-content>

        <!-- Batch action bar -->
        <div class="batch-bar" *ngIf="selection.size > 0">
          <span style="font-weight:500">{{ selection.size }} geselecteerd</span>
          <mat-form-field style="min-width:140px" subscriptSizing="dynamic">
            <mat-label>Klasse instellen</mat-label>
            <mat-select [(value)]="batchClass">
              <mat-option value="">— geen —</mat-option>
              <mat-option value="1">Klasse 1</mat-option>
              <mat-option value="2">Klasse 2</mat-option>
            </mat-select>
          </mat-form-field>
          <button mat-raised-button color="primary" (click)="applyBatchClass()">Toepassen</button>
          <button mat-button (click)="selection.clear()">Deselecteer</button>
        </div>

        <table mat-table [dataSource]="players">

          <!-- Checkbox column -->
          <ng-container matColumnDef="select">
            <th mat-header-cell *matHeaderCellDef style="width:40px">
              <mat-checkbox
                [checked]="allSelected()"
                [indeterminate]="selection.size > 0 && !allSelected()"
                (change)="toggleAll($event.checked)">
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
            <td mat-cell *matCellDef="let p"><strong>{{ p.name }}</strong></td>
          </ng-container>
          <ng-container matColumnDef="city">
            <th mat-header-cell *matHeaderCellDef>Woonplaats</th>
            <td mat-cell *matCellDef="let p">{{ p.city }}</td>
          </ng-container>
          <ng-container matColumnDef="class">
            <th mat-header-cell *matHeaderCellDef style="width:110px">Klasse</th>
            <td mat-cell *matCellDef="let p">
              <mat-select [value]="p.class" (selectionChange)="updateClass(p, $event.value)"
                          style="font-size:14px" panelWidth="">
                <mat-option value="">—</mat-option>
                <mat-option value="1">Klasse 1</mat-option>
                <mat-option value="2">Klasse 2</mat-option>
              </mat-select>
            </td>
          </ng-container>
          <ng-container matColumnDef="buddies">
            <th mat-header-cell *matHeaderCellDef>Speelpartners</th>
            <td mat-cell *matCellDef="let p">
              <mat-chip-set *ngIf="buddyMap[p.id]?.length">
                <mat-chip *ngFor="let bid of buddyMap[p.id]" class="buddy-chip" disableRipple>
                  {{ playerName(bid) }}
                </mat-chip>
              </mat-chip-set>
              <span *ngIf="!buddyMap[p.id]?.length" style="color:#bbb">—</span>
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
            </td>
          </ng-container>
          <tr mat-header-row *matHeaderRowDef="cols"></tr>
          <tr mat-row *matRowDef="let row; columns: cols;"></tr>
        </table>
      </mat-card-content>
    </mat-card>

    <p *ngIf="players.length === 0 && !loading" style="color:#888;text-align:center;padding:32px 0">
      Nog geen spelers geïmporteerd.
    </p>
  `,
})
export class UploadComponent implements OnInit {
  private playerService = inject(PlayerService);
  private snackBar      = inject(MatSnackBar);
  private dialog        = inject(MatDialog);
  private cdr           = inject(ChangeDetectorRef);

  @ViewChild('fileInput') fileInputRef!: ElementRef<HTMLInputElement>;

  selectedFile: File | null = null;
  loading = false;
  players: Player[] = [];
  buddyMap: Record<string, string[]> = {};
  selection = new Set<string>();
  batchClass = '';
  cols = ['select', 'nr', 'name', 'class', 'city', 'buddies', 'actions'];

  ngOnInit(): void { this.loadPlayers(); }

  loadPlayers(): void {
    this.playerService.list().subscribe({
      next: (ps) => {
        this.players = ps;
        this.loadAllBuddies();
      },
      error: () => {},
    });
  }

  loadAllBuddies(): void {
    this.buddyMap = {};
    for (const p of this.players) {
      this.playerService.getBuddies(p.id).subscribe({
        next: (ids) => { this.buddyMap = { ...this.buddyMap, [p.id]: ids }; },
        error: () => {},
      });
    }
  }

  playerName(id: string): string {
    return this.players.find(p => p.id === id)?.name ?? id.slice(0, 8);
  }

  allSelected(): boolean { return this.players.length > 0 && this.selection.size === this.players.length; }

  toggleAll(checked: boolean): void {
    if (checked) this.players.forEach(p => this.selection.add(p.id));
    else this.selection.clear();
  }

  toggleOne(id: string): void {
    if (this.selection.has(id)) this.selection.delete(id);
    else this.selection.add(id);
  }

  updateClass(player: Player, cls: string): void {
    this.playerService.update({ ...player, class: cls }).subscribe({
      next: (p) => { this.players = this.players.map(x => x.id === p.id ? p : x); },
      error: (err) => this.snackBar.open(`Fout: ${err.message}`, 'Sluiten', { duration: 5000 }),
    });
  }

  applyBatchClass(): void {
    const ids = Array.from(this.selection);
    let done = 0;
    for (const id of ids) {
      const player = this.players.find(p => p.id === id);
      if (!player) continue;
      this.playerService.update({ ...player, class: this.batchClass }).subscribe({
        next: (p) => {
          this.players = this.players.map(x => x.id === p.id ? p : x);
          if (++done === ids.length) {
            this.snackBar.open(`Klasse bijgewerkt voor ${done} spelers`, 'OK', { duration: 2000 });
            this.selection.clear();
          }
        },
        error: (err) => this.snackBar.open(`Fout: ${err.message}`, 'Sluiten', { duration: 5000 }),
      });
    }
  }

  onFileSelected(event: Event): void {
    this.selectedFile = (event.target as HTMLInputElement).files?.[0] ?? null;
    this.cdr.detectChanges();
  }

  upload(): void {
    if (!this.selectedFile) return;
    this.loading = true;
    this.playerService.import(this.selectedFile).subscribe({
      next: (res) => {
        this.snackBar.open(`${res.imported} spelers geïmporteerd`, 'OK', { duration: 3000 });
        this.selectedFile = null;
        this.fileInputRef.nativeElement.value = '';
        this.loading = false;
        this.loadPlayers();
      },
      error: (err) => {
        this.snackBar.open(`Fout: ${err.message ?? err.statusText}`, 'Sluiten', { duration: 5000 });
        this.selectedFile = null;
        this.fileInputRef.nativeElement.value = '';
        this.loading = false;
      },
    });
  }

  openEdit(player: Player): void {
    const ref = this.dialog.open(PlayerEditDialogComponent, { data: player });
    ref.afterClosed().subscribe((updated: Player | undefined) => {
      if (!updated) return;
      this.playerService.update(updated).subscribe({
        next: (p) => {
          this.players = this.players.map(x => x.id === p.id ? p : x);
          this.snackBar.open('Speler opgeslagen', 'OK', { duration: 2000 });
        },
        error: (err) => this.snackBar.open(`Fout: ${err.message}`, 'Sluiten', { duration: 5000 }),
      });
    });
  }

  openBuddies(player: Player): void {
    const currentBuddyIds = this.buddyMap[player.id] ?? [];
    const ref = this.dialog.open(BuddyDialogComponent, {
      data: {
        player,
        others: this.players.filter(p => p.id !== player.id),
        currentBuddyIds,
      },
    });
    ref.afterClosed().subscribe((buddyIds: string[] | undefined) => {
      if (buddyIds === undefined) return;
      this.playerService.setBuddies(player.id, buddyIds).subscribe({
        next: () => {
          this.buddyMap = { ...this.buddyMap, [player.id]: buddyIds };
          this.snackBar.open('Speelpartners opgeslagen', 'OK', { duration: 2000 });
        },
        error: (err) => this.snackBar.open(`Fout: ${err.message}`, 'Sluiten', { duration: 5000 }),
      });
    });
  }
}
