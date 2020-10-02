let user = Cypress.env("BASICAUTH_USER") ? Cypress.env("BASICAUTH_USER") : "";
let pass = Cypress.env("BASICAUTH_PASS") ? Cypress.env("BASICAUTH_PASS") : "";

async function getConfirmationEmail() {
  var s = "";
  await cy
    .readFile("/tmp/sfd_confirmación_de_correo_electrónico.txt")
    .then((str) => {
      s = str;
    });
  return s.match(/https*\:\/\/.*/)[0];
}

describe("Registration workflow", function () {
  before(() => {
    // runs once before all tests in the block
    cy.clearCookies();
  });
  beforeEach(() => {
    Cypress.Cookies.preserveOnce("sfd-session-id");
  });
  it("can navigate to site", () => {
    cy.visit("/", {
      auth: { username: user, password: pass },
    });
    cy.get("nav").within(() => {
      cy.get('a[href="/cuentas/registrarse"]').click();
    });
    cy.location(`pathname`).should("eq", "/cuentas/registrarse");
  });
  it(`should display form`, () => {
    cy.get(`input[name="username"`).should("exist");
    cy.get(`input[name="password"`).should("exist");
    cy.get(`input[name="email"`).should("exist");
    cy.get(`input[name="repeat_password"`).should("exist");
    cy.get(`input[name="full_name"`).should("exist");
    cy.get(`input[name="accept_tos"`).should("exist");
  });

  it(`clicking the button should display errors`, () => {
    cy.get(`button[type="submit"]`).click();
  });

  it(`filling it out displays a success message`, () => {
    const id = `new_test_user_${Math.floor(Math.random() * 10000000000000)}`;
    cy.get(`input[name="username"`).type(id);
    cy.get(`input[name="email"`).type(`${id}@email.com`);
    cy.get(`input[name="full_name"`).type(id);
    cy.get(`input[name="password"`).type(`Abcd1234!`);
    cy.get(`input[name="repeat_password"`).type(`Abcd1234!`);
    cy.get(`input[name="accept_tos"`).click();
    cy.get(`button[type="submit"]`).click();
  });

  it(`should give the user a message showing that he has been verified`, () => {
    cy.readFile("/tmp/sfd_confirmación_de_correo_electrónico.txt").then(
      (str) => {
        cy.visit(str.match(/https*\:\/\/.*/)[0]);
        cy.url().should(`match`, /.*\/cuentas\/registrarse\/token.*/);
      }
    );
  });
});
