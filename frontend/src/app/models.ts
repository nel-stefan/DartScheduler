export interface Player {
  id: string;
  nr: string;
  name: string;
  email: string;
  sponsor: string;
  address: string;
  postalCode: string;
  city: string;
  phone: string;
  mobile: string;
  memberSince: string;
  class: string;
}

export interface Match {
  id: string;
  eveningId: string;
  playerA: string;
  playerB: string;
  scoreA: number | null;
  scoreB: number | null;
  played: boolean;
  leg1Winner: string;
  leg1Turns: number;
  leg2Winner: string;
  leg2Turns: number;
  leg3Winner: string;
  leg3Turns: number;
  reportedBy: string;
  rescheduleDate: string;
  secretaryNr: string;
  counterNr: string;
}

export interface Evening {
  id: string;
  number: number;
  date: string;
  isInhaalAvond: boolean;
  matches: Match[];
}

export interface Schedule {
  id: string;
  competitionName: string;
  season: string;
  evenings: Evening[];
  createdAt: string;
}

export interface SeasonSummary {
  id: string;
  competitionName: string;
  season: string;
  createdAt: string;
  eveningCount: number;
}

export interface PlayerStats {
  player: Player;
  played: number;
  wins: number;
  losses: number;
  draws: number;
  pointsFor: number;
  pointsAgainst: number;
}

export interface DutyStats {
  player: Player;
  count: number;
}

export interface GenerateScheduleRequest {
  competitionName: string;
  season: string;
  numEvenings: number;
  startDate: string;
  intervalDays: number;
  inhaalNrs: number[];
  vrijeNrs: number[];
}
