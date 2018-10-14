import {Injectable} from '@angular/core';
import {Observable, of} from "rxjs";
import {HttpClient} from "@angular/common/http";

import {Hero} from './hero'
import {HEROES} from "./mock-heroes";
import {MessageService} from "./message.service";
import {catchError} from "rxjs/operators";

@Injectable({
  providedIn: 'root'
})
export class HeroService {

  // todo url
  private heroesURL: string
  private heroURL: string

  constructor(
    private http: HttpClient,
    private messageService: MessageService
  ) {
  }

  getHeroes(): Observable<Hero[]> {
    this.messageService.add("HeroService: Fetching heroes");
    return this.http.get<Hero[]>(this.heroesURL).pipe(catchError(this.handleError('getHeroes', [])));
  }

  getHeroes2(): Observable<Hero[]> {
    this.messageService.add("HeroService: Fetching heroes");
    return of(HEROES);
  }

  getHero(id: number): Observable<Hero> {
    this.messageService.add(`HeroService: Fetching hero for ${id}`);
    return this.http.get<Hero>(this.heroURL).pipe(catchError(this.handleError('getHero', null)));
  }

  getHero2(id: number): Observable<Hero> {
    this.messageService.add(`HeroService: Fetching hero for ${id}`);
    return of(HEROES.find(hero => hero.id === id))
  }

  private handleError<T>(operation = 'operation', result?: T) {
    return (error: any): Observable<T> => {
      console.error(error);
      this.messageService.add(`${operation} failed: ${error.message}`);
      return of(result as T);
    }
  }
}
