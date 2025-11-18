import { bootstrapApplication } from '@angular/platform-browser';
import { registerLocaleData } from '@angular/common';
import en from '@angular/common/locales/en';
import { appConfig } from './app/app.config';
import { AppComponent } from './app/app.component';

registerLocaleData(en);

bootstrapApplication(AppComponent, appConfig)
  .catch((err) => console.error(err));
