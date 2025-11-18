import { Component, OnInit } from '@angular/core';
import { CommonModule, NgFor, NgIf } from '@angular/common';
import {
  FormBuilder,
  FormGroup,
  ReactiveFormsModule,
  Validators,
} from '@angular/forms';
import { HttpClient } from '@angular/common/http';
import { NzFormModule } from 'ng-zorro-antd/form';
import { NzInputModule } from 'ng-zorro-antd/input';
import { NzSelectModule } from 'ng-zorro-antd/select';
import { NzRadioModule } from 'ng-zorro-antd/radio';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzDatePickerModule } from 'ng-zorro-antd/date-picker';
import {
  NzNotificationModule,
  NzNotificationService,
} from 'ng-zorro-antd/notification';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    NgIf,
    NgFor,
    NzFormModule,
    NzInputModule,
    NzSelectModule,
    NzRadioModule,
    NzButtonModule,
    NzDatePickerModule,
    NzNotificationModule,
  ],
  templateUrl: './app.component.html',
  styleUrl: './app.component.css',
})
export class AppComponent implements OnInit {
  form: FormGroup;
  occupations: string[] = [];
  successMessage = '';
  savedId?: string;

  constructor(
    private fb: FormBuilder,
    private http: HttpClient,
    private notification: NzNotificationService
  ) {
    this.form = this.fb.group({
      firstName: ['', Validators.required],
      lastName: ['', Validators.required],
      email: ['', [Validators.required, Validators.email]],
      phone: ['', [Validators.required, Validators.pattern(/^[0-9]{9,10}$/)]],
      profileBase64: ['', Validators.required],
      birthDay: ['', Validators.required],
      occupation: ['', Validators.required],
      sex: ['Male', Validators.required],
    });
  }

  ngOnInit(): void {
    this.loadOccupations();
  }

  get f() {
    return this.form.controls;
  }

  loadOccupations(): void {
    this.http
      .get<{ items: string[] }>('http://localhost:8081/api/occupations')
      .subscribe({
        next: (res) => (this.occupations = res.items),
        error: () => (this.occupations = []),
      });
  }

  onPhoneInput(event: Event): void {
    const input = event.target as HTMLInputElement;
    const digitsOnly = input.value.replace(/[^0-9]/g, '');
    if (digitsOnly !== input.value) {
      input.value = digitsOnly;
    }
    this.form.patchValue({ phone: digitsOnly }, { emitEvent: false });
  }

  onFileSelected(event: Event): void {
    const input = event.target as HTMLInputElement;
    const file = input.files && input.files[0];
    if (!file) {
      return;
    }

    const reader = new FileReader();
    reader.onload = () => {
      const result = reader.result as string;
      const base64 = result.includes(',') ? result.split(',')[1] : result;
      this.form.patchValue({ profileBase64: base64 });
      this.f['profileBase64'].markAsDirty();
      this.f['profileBase64'].updateValueAndValidity();
    };
    reader.readAsDataURL(file);
  }

  onSave(): void {
    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    const raw = this.form.getRawValue();
    const birth = raw.birthDay ? new Date(raw.birthDay) : null;
    let formattedBirth = '';
    if (birth && !isNaN(birth.getTime())) {
      const day = String(birth.getDate()).padStart(2, '0');
      const month = String(birth.getMonth() + 1).padStart(2, '0');
      const year = birth.getFullYear();
      formattedBirth = `${day}/${month}/${year}`;
    }

    const payload = {
      firstName: raw.firstName,
      lastName: raw.lastName,
      email: raw.email,
      phone: raw.phone,
      profileBase64: raw.profileBase64,
      birthDay: formattedBirth,
      occupation: raw.occupation,
      sex: raw.sex,
    };

    this.http
      .post<{ id: string; message: string }>(
        'http://localhost:8081/api/profiles',
        payload
      )
      .subscribe({
        next: (res) => {
          this.savedId = res.id;
          this.successMessage = `${res.message} : ID ${res.id}`;
          this.notification.success(
            'Save Data',
            `${res.message} : ID ${res.id}`
          );

          this.form.reset({ sex: 'Male' });
          this.form.markAsPristine();
          this.form.markAsUntouched();
        },
        error: () => {
          this.successMessage = 'save data failed';
          this.notification.error('Save Data', 'save data failed');
        },
      });
  }

  onClear(): void {
    this.form.reset({ sex: 'Male' });
    this.successMessage = '';
    this.savedId = undefined;
  }
}
