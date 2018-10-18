import { Component, OnInit } from '@angular/core';

@Component({
  selector: 'app-bot',
  templateUrl: './bot.component.html',
  styleUrls: ['./bot.component.css']
})
export class BotComponent implements OnInit {

  bots: string[];

  constructor() { }

  ngOnInit() {
    this.bots = ["a", "b", "c", "d"];
  }

}
