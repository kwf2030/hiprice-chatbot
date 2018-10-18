import {NgModule} from '@angular/core';
import {RouterModule, Routes} from "@angular/router";
import {BotComponent} from "./bot/bot.component";

const routes: Routes = [
  {path: '', redirectTo: '/bot', pathMatch: 'full'},
  {path: 'bot', component: BotComponent},
]

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule],
})
export class AppRoutingModule {
}
