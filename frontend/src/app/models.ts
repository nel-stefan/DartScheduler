export interface Player {
  id: string;
  name: string;
  email: string;
  sponsor: string;
}

export interface Match {
  id: string;
  eveningId: string;
  playerA: string;
  playerB: string;
  scoreA: number | null;
  scoreB: number | null;
  played: boolean;
}

export interface Evening {
  id: string;
  number: number;
  date: string;
  matches: Match[];
}

export interface Schedule {
  id: string;
  competitionName: string;
  evenings: Evening[];
  createdAt: string;
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

export interface GenerateScheduleRequest {
  competitionName: string;
  numEvenings: number;
  startDate: string;
  intervalDays: number;
}
