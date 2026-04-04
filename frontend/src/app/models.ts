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
  playedDate: string;
  playerA180s: number;
  playerB180s: number;
  playerAHighestFinish: number;
  playerBHighestFinish: number;
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
  oneEighties: number;
  highestFinish: number;
  minTurns: number;
  avgTurns: number;
  avgScorePerTurn: number;
}

export interface DutyMatch {
  eveningNr: number;
  playerANr: string;
  playerAName: string;
  playerBNr: string;
  playerBName: string;
}

export interface DutyStats {
  player: Player;
  count: number;
  secretaryCount: number;
  counterCount: number;
  secretaryMatches: DutyMatch[];
  counterMatches: DutyMatch[];
}

export interface ScheduleInfo {
  players: PlayerInfoItem[];
  evenings: EveningInfoItem[];
  matrix: MatrixCellItem[];
  buddyPairs: BuddyPairItem[];
}

export interface PlayerInfoItem {
  id: string;
  nr: string;
  name: string;
  email: string;
  class: string;
}

export interface EveningInfoItem {
  id: string;
  number: number;
  date: string;
}

export interface MatrixCellItem {
  playerId: string;
  eveningId: string;
  count: number;
}

export interface BuddyPairItem {
  playerAId: string;
  playerANr: string;
  playerAName: string;
  playerBId: string;
  playerBNr: string;
  playerBName: string;
  eveningIds: string[];
  eveningNrs: number[];
}

export interface EveningPlayerStat {
  playerId: string;
  oneEighties: number;
  highestFinish: number;
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
