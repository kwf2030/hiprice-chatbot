import { Component, OnInit } from '@angular/core';
import {Observable, Subject} from "rxjs";
import {Hero} from "../hero";
import {HeroService} from "../hero.service";
import {debounceTime, distinctUntilChanged, switchMap} from "rxjs/operators";

@Component({
  selector: 'app-hero-search',
  templateUrl: './hero-search.component.html',
  styleUrls: ['./hero-search.component.css']
})
export class HeroSearchComponent implements OnInit {

  heroes$: Observable<Hero[]>;
  private searchNames = new Subject<string>();

  constructor(private heroService: HeroService) { }

  ngOnInit() {
    this.heroes$ = this.searchNames.pipe(
      debounceTime(300),
      distinctUntilChanged(),
      switchMap(name => this.heroService.searchHeroes(name))
    );
  }

  search(name: string): void {
    this.searchNames.next(name);
  }

}
