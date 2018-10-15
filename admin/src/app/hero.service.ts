import {Injectable} from '@angular/core';
import {Observable, of} from "rxjs";
import {HttpClient, HttpHeaders} from "@angular/common/http";

import {Hero} from './hero'
import {HEROES} from "./mock-heroes";
import {MessageService} from "./message.service";
import {catchError, tap} from "rxjs/operators";

@Injectable({
  providedIn: 'root'
})
export class HeroService {

  // todo url
  private heroesURL = "api/heroes"
  private heroURL = "api/hero"

  constructor(
    private http: HttpClient,
    private messageService: MessageService
  ) {}

  getHeroes2(): Observable<Hero[]> {
    return this.http.get<Hero[]>(this.heroesURL).pipe(
      tap(_ => this.messageService.add("fetched heroes")),
      catchError(this.handleError('getHeroes', []))
    );
  }

  getHeroes(): Observable<Hero[]> {
    this.messageService.add("HeroService: Fetching heroes");
    return of(HEROES);
  }

  getHero2(id: number): Observable<Hero> {
    const url = `${this.heroURL}/${id}`;
    return this.http.get<Hero>(url).pipe(
      tap(_ => this.messageService.add('fetched hero')),
      catchError(this.handleError('getHero', null))
    );
  }

  getHero(id: number): Observable<Hero> {
    this.messageService.add(`HeroService: Fetching hero for ${id}`);
    return of(HEROES.find(hero => hero.id === id))
  }

  updateHero(hero: Hero): Observable<any> {
    const httpOptions = {
      headers: new HttpHeaders({'Content-Type': 'application/json'})
    }
    return this.http.put(this.heroesURL, hero, httpOptions).pipe(
      tap(_ => this.messageService.add('updateHero')),
      catchError(this.handleError<any>('updateHero'))
    )
  }

  addHero(hero: Hero): Observable<Hero> {
    const httpOptions = {
      headers: new HttpHeaders({'Content-Type': 'application/json'})
    }
    return this.http.post<Hero>(this.heroesURL, hero, httpOptions).pipe(
      tap(hero => this.messageService.add(`add hero: ${hero.id}`)),
      catchError(this.handleError<Hero>('addHero'))
    );
  }

  deleteHero(hero: Hero | number): Observable<Hero> {
    const id = typeof hero === 'number' ? hero : hero.id;
    const url = `${this.heroesURL}/${id}`;
    const httpOptions = {
      headers: new HttpHeaders({'Content-Type': 'application/json'})
    }
    return this.http.delete<Hero>(url, httpOptions).pipe(
      tap(_ => this.messageService.add(`deleted hero id=${id}`)),
      catchError(this.handleError<Hero>('deleteHero'))
    );
  }

  searchHeroes(name: string): Observable<Hero[]> {
    if (!name.trim()) {
      return of([]);
    }
    return this.http.get<Hero[]>(`${this.heroesURL}/?name=${name}`).pipe(
      tap(_ => this.messageService.add(`found heroes matching ${name}`)),
      catchError(this.handleError<Hero[]>('searchHeroes', []))
    );
  }

  private handleError<T>(operation = 'operation', result?: T) {
    return (error: any): Observable<T> => {
      console.error(error);
      this.messageService.add(`${operation} failed: ${error.message}`);
      return of(result as T);
    }
  }
}
