import { Component, inject, OnInit } from '@angular/core';
import { Router } from '@angular/router';

/** Scores worden nu ingevoerd via het dialoogvenster in het schema-overzicht. */
@Component({
  selector: 'app-score-entry',
  imports: [],
  template: '',
})
export class ScoreEntryComponent implements OnInit {
  private router = inject(Router);
  ngOnInit(): void {
    this.router.navigate(['/']);
  }
}
