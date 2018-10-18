import {BrowserModule} from '@angular/platform-browser';
import {NgModule} from '@angular/core';
import {FormsModule} from '@angular/forms';
import {HttpClientModule} from "@angular/common/http";
import {BrowserAnimationsModule} from '@angular/platform-browser/animations';
import {registerLocaleData} from '@angular/common';
import zh from '@angular/common/locales/zh';

import {NgZorroAntdModule, NZ_I18N, zh_CN} from 'ng-zorro-antd';

import {AppComponent} from './app.component';
import {AppRoutingModule} from './app-routing.module';
import {BotComponent} from './bot/bot.component';

registerLocaleData(zh);

@NgModule({
  declarations: [
    AppComponent,
    BotComponent,
  ],
  imports: [
    BrowserModule,
    FormsModule,
    HttpClientModule,
    BrowserAnimationsModule,
    NgZorroAntdModule,
    AppRoutingModule
  ],
  providers: [{provide: NZ_I18N, useValue: zh_CN}],
  bootstrap: [AppComponent]
})
export class AppModule {
}
