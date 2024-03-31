export interface Costume {
  school: string | null;
  racing: string | null;
  original: string | null;
  staringFuture: string | null;
}

export interface Profile {
  name: string | null;
  birthday: string | null;
  url: string;
  costume: Costume;
}
